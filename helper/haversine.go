package helper

import "math"

func Haversine(lat1, lon1, lat2, lon2 float64) float64 {

	dlat := (lat2 - lat1) * (math.Pi / 180.0)
	dlon := (lon2 - lon1) * (math.Pi / 180.0)

	lat1 = (lat1) * (math.Pi) / 180.0
	lat2 = (lat2) * (math.Pi) / 180.0

	a := math.Pow(math.Sin(dlat/2), 2) + math.Pow(math.Sin(dlon/2), 2)*
		math.Cos(lat1)*math.Cos(lat2)

	rad := 6371.0
	c := 2.0 * math.Asin(math.Sqrt(a))
	return rad * c
}
