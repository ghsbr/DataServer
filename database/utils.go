package database

import (
	"encoding/binary"
	"math"
	"time"
)
// Ottieni il nome della relativa tabella dalla longitudine.
func getIndexFromLongitude(long float64) int64 {
	posIdx := int64(math.Trunc(long))
	return posIdx - posIdx%longPerTable
}
//Tronca un orario al giorno inerente, cancellando i dati relativi alle ore.
func truncateTime(timestamp int64) int64 {
	return time.Unix(timestamp, 0).Truncate(time.Hour * 24).Unix()
}
//Errore ritornato in caso di mancanza di stazione.
type NotFoundError struct {
	msg string
}

func (e NotFoundError) Error() string {
	return e.msg
}
//Ottieni la rappresentazione binaria di un double.
//Corretto solo su architetture Little Endian.
func floatToBytes(num float64) [8]byte {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(num))
	return buf
}
//Ritorna il modulo del numero
func abs(num int64) int64 {
	if num < 0 {
		return -num
	}
	return num
}
