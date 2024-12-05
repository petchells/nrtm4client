// Package jsonseq provides functions for splitting a jsonseq file into records
//
// A jsonseq record is simply the bytes between the record markers -- it's up to
// you to unmarshall them to the JSON types you expect.
package jsonseq
