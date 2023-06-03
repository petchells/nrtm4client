package persist

type Repository interface {
	InitializeConnectionPool(dbUrl string)
	GetState() (NRTMState, error)
}
