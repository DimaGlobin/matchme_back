package utils

import (
	"hash/fnv"
	"math"
	"time"

	"github.com/lib/pq"
)

func Contains(arr pq.Int64Array, target int64) bool {
	for _, val := range arr {
		if val == target {
			return true
		}
	}
	return false
}

func Hash(b []byte) (uint32, error) {
	h := fnv.New32a()
	_, err := h.Write(b)
	return h.Sum32(), err
}

func VincentyDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const (
		a       = 6378137.0         // Экваториальный радиус Земли
		f       = 1 / 298.257223563 // Сжатие Земли
		b       = (1 - f) * a       // Полярный радиус Земли
		maxIter = 200
		eps     = 1e-12
	)

	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	U1 := math.Atan((1 - f) * math.Tan(lat1Rad))
	U2 := math.Atan((1 - f) * math.Tan(lat2Rad))
	L := lon2Rad - lon1Rad

	lambda := L
	sinU1 := math.Sin(U1)
	cosU1 := math.Cos(U1)
	sinU2 := math.Sin(U2)
	cosU2 := math.Cos(U2)

	iter := 0
	var sinLambda, cosLambda, sinSigma, cosSigma, sigma, sinAlpha, cosSqAlpha, cos2SigmaM float64

	for {
		sinLambda = math.Sin(lambda)
		cosLambda = math.Cos(lambda)
		sinSigma = math.Sqrt((cosU2*sinLambda)*(cosU2*sinLambda) + (cosU1*sinU2-sinU1*cosU2*cosLambda)*(cosU1*sinU2-sinU1*cosU2*cosLambda))
		cosSigma = sinU1*sinU2 + cosU1*cosU2*cosLambda
		sigma = math.Atan2(sinSigma, cosSigma)
		sinAlpha = cosU1 * cosU2 * sinLambda / sinSigma
		cosSqAlpha = 1 - sinAlpha*sinAlpha
		cos2SigmaM = cosSigma - 2*sinU1*sinU2/cosSqAlpha

		C := f / 16 * cosSqAlpha * (4 + f*(4-3*cosSqAlpha))
		lambdaPrev := lambda
		lambda = L + (1-C)*f*sinAlpha*(sigma+C*sinSigma*(cos2SigmaM+C*cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)))
		iter++

		if math.Abs(lambda-lambdaPrev) <= eps || iter >= maxIter {
			break
		}
	}

	uSq := cosSqAlpha * (a*a - b*b) / (b * b)
	A := 1 + uSq/16384*(4096+uSq*(-768+uSq*(320-175*uSq)))
	B := uSq / 1024 * (256 + uSq*(-128+uSq*(74-47*uSq)))
	deltaSigma := B * sinSigma * (cos2SigmaM + B/4*(cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)-B/6*cos2SigmaM*(-3+4*sinSigma*sinSigma)*(-3+4*cos2SigmaM*cos2SigmaM)))

	distance := b * A * (sigma - deltaSigma)

	return distance
}

func RemoveNumberFromArray(arr pq.Int64Array, num uint32) pq.Int64Array {
	index := -1
	for i, val := range arr {
		if val == int64(num) {
			index = i
			break
		}
	}

	if index != -1 {
		// Удаляем элемент по индексу index
		arr = append(arr[:index], arr[index+1:]...)
		return arr
	}

	return nil
}

func ParseDate(dateStr string) (time.Time, error) {
	layout := "02.01.2006"
	parsedDate, err := time.Parse(layout, dateStr)
	if err != nil {
		return time.Time{}, err
	}

	return parsedDate, nil
}

func CalculateAge(birthDate time.Time) int {
	today := time.Now()
	age := today.Year() - birthDate.Year()
	if today.YearDay() < birthDate.YearDay() {
		age--
	}
	return age
}
