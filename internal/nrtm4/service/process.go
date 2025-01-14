package service

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/petchells/nrtm4client/internal/nrtm4/jsonseq"
	"github.com/petchells/nrtm4client/internal/nrtm4/persist"
	"github.com/petchells/nrtm4client/internal/nrtm4/pg/db"
	"github.com/petchells/nrtm4client/internal/nrtm4/rpsl"
	"github.com/petchells/nrtm4client/internal/nrtm4/util"
)

var (
	// Protocol errors

	// ErrNRTMVersionMismatch nrtm version is not 4
	ErrNRTMVersionMismatch = errors.New("nrtm version is not 4")
	// ErrNRTMSourceMismatch session id does not match source
	ErrNRTMSourceMismatch = errors.New("session id does not match source")
	// ErrNRTMSourceNameMismatch source name does not match source
	ErrNRTMSourceNameMismatch = errors.New("source name does not match source")
	// ErrNRTMFileVersionMismatch file version does not match its reference
	ErrNRTMFileVersionMismatch = errors.New("file version does not match its reference")
	// ErrNRTMFileVersionInconsistency version is lower than source
	ErrNRTMFileVersionInconsistency = errors.New("version is lower than source")
	// ErrNRTMNoDeltasInNotification the NRTM server published a notification file with no deltas
	ErrNRTMNoDeltasInNotification = errors.New("no deltas listed in notification file")
	// ErrNRTMNotificationDeltaSequenceBroken the NRTM server has an incontiguous list of delta version
	ErrNRTMNotificationDeltaSequenceBroken = errors.New("server has incontiguous list of delta versions")
	// ErrNRTMNotificationVersionDoesNotMatchDelta the highest delta version is not the notification version
	ErrNRTMNotificationVersionDoesNotMatchDelta = errors.New("highest delta version is not the notification version")
	// ErrNRTMDuplicateDeltaVersion the highest delta version is not the notification version
	ErrNRTMDuplicateDeltaVersion = errors.New("notification file published a duplicate delta file")

	// Repo errors

	// ErrNextConsecutiveDeltaUnavaliable cannot find the next consecutive delta to apply to our repo
	ErrNextConsecutiveDeltaUnavaliable = errors.New("repository is too old to update from the server")
	// ErrSourceAlreadyExists a source with the given label already exists
	ErrSourceAlreadyExists = errors.New("a source with the given label already exists")

	fileWriteBufferLength = 1024 * 8
	rpslInsertBatchSize   = 1000
)

// AppConfig application configuration object
type AppConfig struct {
	NRTMFilePath     string
	PgDatabaseURL    string
	BoltDatabasePath string
}

// NewNRTMProcessor injects repo and client into service and return a new instance
func NewNRTMProcessor(config AppConfig, repo persist.Repository, client Client) NRTMProcessor {
	return NRTMProcessor{
		config: config,
		repo:   repo,
		client: client,
	}
}

// NRTMProcessor orchestration for functions the client implements
type NRTMProcessor struct {
	config AppConfig
	repo   persist.Repository
	client Client
}

var labelRe = regexp.MustCompile("^[A-Za-z0-9 ._-]*[A-Za-z0-9][A-Za-z0-9 ._-]*$")

// Connect stores details about a connection
func (p NRTMProcessor) Connect(notificationURL string, label string) error {
	label = strings.TrimSpace(label)
	if len(label) > 0 && !labelRe.MatchString(label) {
		return errors.New("Label is not valid")
	}
	logger.Info("Fetching notification")
	fm := fileManager{p.client}
	notification, errs := fm.downloadNotificationFile(notificationURL)
	if len(errs) > 0 {
		return errors.New("download error(s): " + errs[0].Error())
	}
	ds := NrtmDataService{Repository: p.repo}
	if ds.getSourceByURLAndLabel(notificationURL, label) != nil {
		return errors.New("source already exists")
	}
	err := fm.ensureDirectoryExists(p.config.NRTMFilePath)
	if err != nil {
		return err
	}
	// Download snapshot
	logger.Info("Fetching snapshot file...")
	snapshotFile, err := fm.fetchFileAndCheckHash(notification.SnapshotRef, p.config.NRTMFilePath)
	if err != nil {
		return err
	}
	logger.Info("Snapshot file downloaded")
	defer snapshotFile.Close()

	logger.Info("Saving new source", "source", notification.Source)
	source := persist.NewNRTMSource(notification, label, notificationURL)
	if source, err = ds.saveNewSource(source, notification); err != nil {
		logger.Error("There was a problem saving the source. Remove it and restart sync", "error", err)
		return err
	}
	logger.Info("Inserting snapshot objects")
	if err := fm.readJSONSeqRecords(snapshotFile, snapshotObjectInsertFunc(p.repo, source, notification)); err != io.EOF {
		logger.Error("Invalid snapshot. Remove Source and restart sync", "error", err)
		return err
	}
	return p.syncDeltas(notification, source)
}

// Update brings the local mirror up to date
func (p NRTMProcessor) Update(sourceName string, label string) error {
	ds := NrtmDataService{Repository: p.repo}
	source := ds.getSourceByNameAndLabel(sourceName, label)
	if source == nil {
		logger.Warn("No source with given name and label", "name", sourceName, "label", label)
		return errors.New("no source found")
	}
	fm := fileManager{p.client}
	notification, errs := fm.downloadNotificationFile(source.NotificationURL)
	if len(errs) > 0 {
		for _, e := range errs {
			logger.Error("Problem downloading notification file", "error", e)
		}
		return errors.New("problem downloading notification file")
	}
	if notification.SessionID != source.SessionID {
		return errors.New("server has a new mirror session")
	}
	if notification.Version < source.Version {
		return errors.New("server has old version")
	}
	if notification.Version == source.Version {
		logger.Info("Already at latest version")
		return nil
	}
	return p.syncDeltas(notification, *source)
}

// ListSources shows all sources
func (p NRTMProcessor) ListSources() ([]persist.NRTMSourceDetails, error) {
	ds := NrtmDataService{Repository: p.repo}
	sources, err := ds.getSources()
	deets := []persist.NRTMSourceDetails{}
	if err != nil {
		return deets, err
	}
	for _, src := range sources {
		to := src.Version
		from := src.Version - 99
		if src.Version <= 99 {
			from = 1
		}
		notifs, err := p.repo.GetNotificationHistory(src, from, to)
		if err != nil {
			return deets, err
		}
		deets = append(deets, persist.NRTMSourceDetails{NRTMSource: src, Notifications: notifs})
	}
	return deets, nil
}

// ReplaceLabel replaces a label name
func (p NRTMProcessor) ReplaceLabel(src, fromLabel, toLabel string) (*persist.NRTMSource, error) {
	ds := NrtmDataService{Repository: p.repo}
	target := ds.getSourceByNameAndLabel(src, fromLabel)
	if target == nil {
		return nil, errors.New("cannot find source with given name and label")
	}
	possDupe := ds.getSourceByNameAndLabel(src, toLabel)
	if possDupe != nil {
		return nil, ErrSourceAlreadyExists
	}
	target.Label = toLabel
	return target, db.WithTransaction(func(tx pgx.Tx) error {
		return db.Update(tx, target)
	})
}

func (p NRTMProcessor) syncDeltas(notification persist.NotificationJSON, source persist.NRTMSource) error {
	deltaRefs, err := findUpdates(notification, source)
	if err != nil {
		return err
	}
	sort.Sort(fileRefsByVersion(deltaRefs))
	fm := fileManager{p.client}
	for _, deltaRef := range deltaRefs {
		logger.Info("Processing delta", "delta", deltaRef.Version, "url", deltaRef.URL)
		file, err := fm.fetchFileAndCheckHash(deltaRef, p.config.NRTMFilePath)
		if err != nil {
			return err
		}
		defer file.Close()
		if err := fm.readJSONSeqRecords(file, applyDeltaFunc(p.repo, source, notification, deltaRef)); err != io.EOF {
			logger.Warn("Failed to apply delta", "source", source, "error", err)
			return err
		}
	}
	logger.Info("Finished syncing deltas")
	return nil
}

func applyDeltaFunc(repo persist.Repository, source persist.NRTMSource, notification persist.NotificationJSON, deltaRef persist.FileRefJSON) jsonseq.RecordReaderFunc {
	var header *persist.DeltaFileJSON
	return func(bytes []byte, err error) error {
		if err == &persist.ErrNoEntity {
			logger.Warn("error empty JSON", "error", err)
			return err
		}
		if err == nil || err == io.EOF {
			if header == nil {
				deltaHeader := new(persist.DeltaFileJSON)
				if err = json.Unmarshal(bytes, deltaHeader); err != nil {
					return err
				}
				if err = validateDeltaHeader(deltaHeader.NrtmFileJSON, source, deltaRef); err != nil {
					return err
				}
				header = deltaHeader
				source.Version = deltaRef.Version
				_, err = repo.SaveSource(source, notification)
				return err
			}
			delta := new(persist.DeltaJSON)
			if err = json.Unmarshal(bytes, delta); err != nil {
				return err
			}
			if delta.Action == persist.DeltaAddModifyAction {
				rpsl, err := rpsl.ParseString(*delta.Object)
				if err != nil {
					return err
				}
				err = repo.AddModifyObject(source, rpsl, header.NrtmFileJSON)
				if err != nil {
					logger.Error("Delta AddModifyObject failed", "rpsl", rpsl, "error", err)
					return err
				}
			} else if delta.Action == persist.DeltaDeleteAction {
				repo.DeleteObject(source, *delta.ObjectClass, *delta.PrimaryKey, header.NrtmFileJSON)
			} else {
				return errors.New("no delta action available: " + delta.Action)
			}
			return nil
		}
		return err
	}
}

// RPSLObjectList an ummutable list of objects
type RPSLObjectList struct {
	mu      sync.Mutex
	objects []rpsl.Rpsl
}

// NewRPSLObjectList returns an initialized RPSLObjectList
func NewRPSLObjectList() RPSLObjectList {
	return RPSLObjectList{objects: make([]rpsl.Rpsl, 0, rpslInsertBatchSize*2)}
}

// Add adds an object the list
func (l *RPSLObjectList) Add(obj rpsl.Rpsl) {
	l.mu.Lock()
	l.objects = append(l.objects, obj)
	l.mu.Unlock()
}

// GetBatch will return a slice of objects only if 'size' are available. They are removed from the list
func (l *RPSLObjectList) GetBatch(size int) []rpsl.Rpsl {
	res := []rpsl.Rpsl{}
	l.mu.Lock()
	if len(l.objects) >= size {
		res = l.objects[:size]
		l.objects = l.objects[size:]
	}
	l.mu.Unlock()
	return res
}

// GetAll returns all RPSL objects and empties the internal list.
func (l *RPSLObjectList) GetAll() []rpsl.Rpsl {
	l.mu.Lock()
	res := l.objects
	l.objects = []rpsl.Rpsl{}
	l.mu.Unlock()
	return res
}

type rpslObjectParser struct{}

type rpslParserPool struct {
	Parsers chan rpslObjectParser
}

func newParserPool(limit int) *rpslParserPool {
	pool := rpslParserPool{}
	pool.Parsers = make(chan rpslObjectParser, limit)
	for range limit {
		pool.Parsers <- rpslObjectParser{}
	}
	return &pool
}

func (pool *rpslParserPool) Acquire() rpslObjectParser {
	return <-pool.Parsers
}

func (pool *rpslParserPool) Release(p rpslObjectParser) {
	pool.Parsers <- p
}

func (pool *rpslParserPool) Close() {
	close(pool.Parsers)
}

func (p *rpslObjectParser) bytesToRPSL(bytes []byte) *rpsl.Rpsl {
	so := new(persist.SnapshotObjectJSON)
	if err := json.Unmarshal(bytes, so); err != nil {
		logger.Warn("Failed to unmarshal RPSL string from", "so.Object", so.Object, "error", err)
		return nil
	}
	rpsl, err := rpsl.ParseString(so.Object)
	if err != nil {
		logger.Warn("Failed to parse rpsl.Rpsl from", "so.Object", so.Object, "error", err)
	}
	return &rpsl

}

func snapshotObjectInsertFunc(repo persist.Repository, source persist.NRTMSource, notification persist.NotificationJSON) jsonseq.RecordReaderFunc {

	var snapshotHeader *persist.SnapshotFileJSON
	var wg sync.WaitGroup

	objectList := NewRPSLObjectList()
	successfulObjects := 0
	failedObjects := 0

	parserPool := newParserPool(4)
	incrementCounters := func(res *rpsl.Rpsl) {
		if obj := res; obj != nil {
			objectList.Add(*obj)
			successfulObjects++
		} else {
			failedObjects++
		}
		rpslObjects := objectList.GetBatch(rpslInsertBatchSize)
		if len(rpslObjects) > 0 {
			err := repo.SaveSnapshotObjects(source, rpslObjects, snapshotHeader.NrtmFileJSON)
			if err != nil {
				log.Fatalln("Error saving snapshot object", err)
			}
		}
	}

	return func(bytes []byte, err error) error {
		if err == &persist.ErrNoEntity {
			logger.Warn("empty JSON record", "error", err)
			return nil
		}
		if err == io.EOF {
			// Expected error reading to end of snapshot objects
			parser := parserPool.Acquire()
			incrementCounters(parser.bytesToRPSL(bytes))
			parserPool.Release(parser)
			wg.Wait()
			parserPool.Close()

			rpslObjects := objectList.GetAll()
			err = repo.SaveSnapshotObjects(source, rpslObjects, snapshotHeader.NrtmFileJSON)
			if err != nil {
				return err
			}
			source.Version = snapshotHeader.Version
			_, err = repo.SaveSource(source, notification)
			return err
		} else if err != nil {
			logger.Warn("error reading jsonseq records.", "error", err)
			return err
		} else if successfulObjects == 0 {
			// First record is the Snapshot header
			successfulObjects++
			sf := new(persist.SnapshotFileJSON)
			if err = json.Unmarshal(bytes, sf); err != nil {
				logger.Warn("error unmarshalling JSON. Expected SnapshotFile", "error", err, "numFailures", failedObjects)
				return err
			}
			if sf.Version != notification.SnapshotRef.Version {
				return ErrNRTMFileVersionMismatch
			}
			snapshotHeader = sf
			return nil
		} else {
			// Subsequent records are objects
			parser := parserPool.Acquire()
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer parserPool.Release(parser)
				incrementCounters(parser.bytesToRPSL(bytes))
			}()
			return nil
		}
	}
}

func validateDeltaHeader(file persist.NrtmFileJSON, source persist.NRTMSource, deltaRef persist.FileRefJSON) error {
	if file.NrtmVersion != 4 {
		return ErrNRTMVersionMismatch
	}
	if file.SessionID != source.SessionID {
		return ErrNRTMSourceMismatch
	}
	if file.Source != source.Source {
		return ErrNRTMSourceNameMismatch
	}
	if file.Version != deltaRef.Version {
		return ErrNRTMFileVersionMismatch
	}
	if file.Version < source.Version {
		return ErrNRTMFileVersionInconsistency
	}
	return nil
}

func findUpdates(notification persist.NotificationJSON, source persist.NRTMSource) ([]persist.FileRefJSON, error) {

	if notification.DeltaRefs == nil || len(*notification.DeltaRefs) == 0 {
		return nil, ErrNRTMNoDeltasInNotification
	}

	deltaRefs := []persist.FileRefJSON{}
	versions := make([]uint32, len(*notification.DeltaRefs))
	for i, deltaRef := range *notification.DeltaRefs {
		versions[i] = deltaRef.Version
		if deltaRef.Version > source.Version {
			deltaRefs = append(deltaRefs, deltaRef)
		}
	}
	versionSet := util.NewSet(versions...)
	if len(versionSet) != len(versions) {
		logger.Error("Duplicate delta version found in notification file", "source", notification.Source, "url", source.NotificationURL)
		return nil, ErrNRTMDuplicateDeltaVersion
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i] < versions[j]
	})
	lo := versions[0]
	hi := versions[len(versions)-1]
	if hi != notification.Version {
		return nil, ErrNRTMNotificationVersionDoesNotMatchDelta
	}
	for i := 0; i < len(versions)-1; i++ {
		if versions[i]+1 != versions[i+1] {
			logger.Error("Delta version is missing from the notification file", "version", versions[i]+1, "source", notification.Source, "url", source.NotificationURL)
			return nil, ErrNRTMNotificationDeltaSequenceBroken
		}
	}
	if source.Version+1 < lo {
		return nil, ErrNextConsecutiveDeltaUnavaliable
	}
	// source.Version == hi // can never happen irl, coz callling fn has already checked Version, and we checked 'hi' above
	logger.Info("Found deltas", "source", notification.Source, "numdeltas", len(deltaRefs))
	return deltaRefs, nil
}
