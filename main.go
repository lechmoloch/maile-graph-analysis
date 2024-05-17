/*
- Тетта целочисленная?
- ОДЗ тетты?
- Какой шаг брать? Если целочисленный, то 1?
- Частота от 0 до 1?
- В расчет берём только успешные выполнения?
*/
package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
)

type graphPoint struct {
	theta     float64
	frequence float64
}

type thetaResultPair struct {
	theta  float64
	result bool
}

func parseThetaResultPairs() []thetaResultPair {
	file, err := os.Open("Data/data.csv")
	if err != nil {
		panic(err)
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 2

	var pairs []thetaResultPair
	for {
		record, e := reader.Read()
		if e != nil {
			fmt.Println(e)
			break
		}

		resultValue, err1 := strconv.Atoi(record[0])
		if err1 != nil {
			panic(err1)
		}

		thetaValue, err2 := strconv.Atoi(record[1])
		if err2 != nil {
			panic(err2)
		}

		var pair thetaResultPair
		pair.result = resultValue != 0
		pair.theta = float64(thetaValue)

		pairs = append(pairs, pair)
	}

	return pairs
}

func minMaxTheta(pairs []thetaResultPair) (float64, float64) {
	min := 100.0
	max := -100.0

	for _, value := range pairs {
		if value.theta < min {
			min = value.theta
		}

		if value.theta > max {
			max = value.theta
		}
	}

	return min, max
}

func birnbaum(c float64, delta float64, theta float64) float64 {
	exponent := math.Exp(1.71 * (theta - delta))
	return c + (1-c)*(exponent/(1+exponent))
}

func goodAnswersAmount(pairs []thetaResultPair) int {
	amount := 0
	for _, value := range pairs {
		if value.result {
			amount++
		}
	}

	return amount
}

func writeToCSV(fileName string, points []graphPoint) {
	file, err := os.Create("Output/" + fileName + ".csv")
	if err != nil {
		panic(err)
	}

	writer := csv.NewWriter(file)
	for _, value := range points {
		var record [2]string = [2]string{strconv.FormatFloat(value.theta, 'f', -1, 64), strconv.FormatFloat(value.frequence, 'f', -1, 64)}
		if err := writer.Write(record[:]); err != nil {
			panic(err)
		}
	}

	writer.Flush()
}

func main() {
	delta := 2.0
	c := 0.001388889
	pairs := parseThetaResultPairs()
	min, max := minMaxTheta(pairs)
	step := 1.0
	var pockets []float64

	for i := min; i <= max; i += step {
		pockets = append(pockets, i)
	}

	var pocketPairs [][]thetaResultPair
	for _, currentPocket := range pockets {
		var currentPairs []thetaResultPair

		for _, currentPair := range pairs {
			if currentPair.theta == currentPocket {
				currentPairs = append(currentPairs, currentPair)
			}
		}

		pocketPairs = append(pocketPairs, currentPairs)
	}

	var birnbaumPoints []graphPoint
	for _, currentPocket := range pockets {
		var currentPoint graphPoint
		currentPoint.theta = currentPocket
		currentPoint.frequence = birnbaum(c, delta, currentPocket)
		birnbaumPoints = append(birnbaumPoints, currentPoint)
	}

	// ДЕБАГ
	fmt.Println("\nПРЕДПОЛАГАЕМЫЕ РЕЗУЛЬТАТЫ (БИРНБАУМ)")
	for _, currentPoint := range birnbaumPoints {
		fmt.Printf("%f; %f\n", currentPoint.theta, currentPoint.frequence)
	}

	var actualPoints []graphPoint
	for _, currentPocketPairs := range pocketPairs {
		theta := currentPocketPairs[0].theta
		frequence := float64(goodAnswersAmount(currentPocketPairs)) / float64(len(currentPocketPairs))

		var currentPoint graphPoint = graphPoint{theta: theta, frequence: frequence}
		actualPoints = append(actualPoints, currentPoint)
	}

	// ДЕБАГ
	fmt.Println("\nФАКТИЧЕСКИЕ РЕЗУЛЬТАТЫ")
	for _, currentPoint := range actualPoints {
		fmt.Printf("%f; %f\n", currentPoint.theta, currentPoint.frequence)
	}

	fmt.Println("\nДЕБАГ ИНТЕРВАЛОВ")
	var positiveConfidenceIntervalPoints, negativeConfidenceIntervalPoints []graphPoint
	for _, currentPocketPairs := range pocketPairs {
		studentsAmount := float64(len(currentPocketPairs))
		frequence := float64(goodAnswersAmount(currentPocketPairs)) / float64(len(currentPocketPairs))
		sigma := math.Sqrt(studentsAmount * frequence * (1 - frequence))

		theta := currentPocketPairs[0].theta
		var pointPositive, pointNegative graphPoint
		pointPositive = graphPoint{theta: theta, frequence: (birnbaum(c, delta, theta) + sigma)}
		pointNegative = graphPoint{theta: theta, frequence: (birnbaum(c, delta, theta) - sigma)}
		positiveConfidenceIntervalPoints = append(positiveConfidenceIntervalPoints, pointPositive)
		negativeConfidenceIntervalPoints = append(negativeConfidenceIntervalPoints, pointNegative)

		fmt.Printf("%f %f %f %f\n", studentsAmount, frequence, sigma, theta)
	}

	// ДЕБАГ
	fmt.Println("\nИНТЕРВАЛЫ")
	for _, currentPoint := range positiveConfidenceIntervalPoints {
		fmt.Printf("%f; %f\n", currentPoint.theta, currentPoint.frequence)
	}
	for _, currentPoint := range negativeConfidenceIntervalPoints {
		fmt.Printf("%f; %f\n", currentPoint.theta, currentPoint.frequence)
	}

	writeToCSV("birnbaumPoints", birnbaumPoints)
	writeToCSV("actualPoints", actualPoints)
	writeToCSV("positiveConfidenceIntervalPoints", positiveConfidenceIntervalPoints)
	writeToCSV("negativeConfidenceIntervalPoints", negativeConfidenceIntervalPoints)
}