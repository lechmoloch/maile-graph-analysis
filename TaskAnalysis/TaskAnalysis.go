package TaskAnalysis

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
)

// * БЛОК СТРУКТУР И МЕТОДОВ

// Точка графика
type graphPoint struct {
	theta     float64
	frequence float64
}

// Пара результат - уровень подготовленности
type thetaResultPair struct {
	theta  float64
	result bool
}

type TaskAnalyzer struct {
	actualDifficulty    float64
	guessingProbability float64
	resultPairs         []thetaResultPair
}

// Конструктор TaskAnalyzer
func NewTaskAnalyzer(
	actualDifficulty float64,
	guessingProbability float64,
	resultDataFilePath string,
) TaskAnalyzer {
	var analyzer TaskAnalyzer
	analyzer.actualDifficulty = actualDifficulty
	analyzer.guessingProbability = guessingProbability
	analyzer.resultPairs = parseThetaResultPairs(resultDataFilePath)

	return analyzer
}

// Производит csv-файл с точками для построения графика
func (analyzer TaskAnalyzer) WriteGraphData(outputFilePath string) {
	file, err := os.Create(outputFilePath)
	if err != nil {
		panic(err)
	}

	pocketPairs := analyzer.getResultPocketPairs()
	pointsActual := analyzer.calculateActualPoints(pocketPairs)
	pointsBirnbaum := analyzer.calculateBirnbaumPoints(pocketPairs)
	pointsPositiveInterval, pointsNegativeInterval := analyzer.calculateConfidenceIntervalsPoints(pocketPairs)

	writer := csv.NewWriter(file)
	for i := 0; i < len(pointsActual); i++ {
		var record [8]string = [8]string{
			strconv.FormatFloat(pointsActual[i].theta, 'f', -1, 64),
			strconv.FormatFloat(pointsActual[i].frequence, 'f', -1, 64),
			strconv.FormatFloat(pointsBirnbaum[i].theta, 'f', -1, 64),
			strconv.FormatFloat(pointsBirnbaum[i].frequence, 'f', -1, 64),
			strconv.FormatFloat(pointsPositiveInterval[i].theta, 'f', -1, 64),
			strconv.FormatFloat(pointsPositiveInterval[i].frequence, 'f', -1, 64),
			strconv.FormatFloat(pointsNegativeInterval[i].theta, 'f', -1, 64),
			strconv.FormatFloat(pointsNegativeInterval[i].frequence, 'f', -1, 64),
		}
		if err := writer.Write(record[:]); err != nil {
			panic(err)
		}
	}

	writer.Flush()
}

// Создание карманов
// TODO - переделать механизм формирования карманов в менее дуболомный
func (analyzer TaskAnalyzer) getResultPocketPairs() [][]thetaResultPair {
	min, max := minMaxTheta(analyzer.resultPairs)
	step := 1.0
	
	var pockets []float64
	for i := min; i <= max; i += step {
		pockets = append(pockets, i)
	}

	var pocketPairs [][]thetaResultPair
	for _, currentPocket := range pockets {
		var currentPairs []thetaResultPair

		for _, currentPair := range analyzer.resultPairs {
			if currentPair.theta == currentPocket {
				currentPairs = append(currentPairs, currentPair)
			}
		}
		pocketPairs = append(pocketPairs, currentPairs)
	}

	return pocketPairs
}

// Вычисление точек фактических результатов
func (analyzer TaskAnalyzer) calculateActualPoints(pocketPairs [][]thetaResultPair) []graphPoint {
	var actualPoints []graphPoint
	for _, currentPocketPairs := range pocketPairs {
		theta := currentPocketPairs[0].theta
		frequence := float64(goodAnswersAmount(currentPocketPairs)) / float64(len(currentPocketPairs))

		var currentPoint graphPoint = graphPoint{theta, frequence}
		actualPoints = append(actualPoints, currentPoint)
	}

	return actualPoints
}

// Вычисление точек предполагаемых результатов
func (analyzer TaskAnalyzer) calculateBirnbaumPoints(pocketPairs [][]thetaResultPair) []graphPoint {
	var birnbaumPoints []graphPoint
	for _, currentPocket := range pocketPairs {
		var currentPoint graphPoint
		currentPoint.theta = currentPocket[0].theta
		currentPoint.frequence = birnbaum(
			analyzer.guessingProbability, 
			analyzer.actualDifficulty, 
			currentPocket[0].theta,
		)
		birnbaumPoints = append(birnbaumPoints, currentPoint)
	}

	return birnbaumPoints
}

// Вычисление точек доверительных интервалов
func (analyzer TaskAnalyzer) calculateConfidenceIntervalsPoints(pocketPairs [][]thetaResultPair) ([]graphPoint, []graphPoint) {
	var positiveConfidenceIntervalPoints, negativeConfidenceIntervalPoints []graphPoint
	for _, currentPocketPairs := range pocketPairs {
		studentsAmount := float64(len(currentPocketPairs))
		frequence := float64(goodAnswersAmount(currentPocketPairs)) / float64(len(currentPocketPairs))
		sigma := math.Sqrt(studentsAmount * frequence * (1 - frequence))

		theta := currentPocketPairs[0].theta
		var pointPositive, pointNegative graphPoint

		pointPositive = graphPoint{
			theta,
			(birnbaum(analyzer.guessingProbability, analyzer.actualDifficulty, theta) + sigma),
		}
		pointNegative = graphPoint{
			theta,
			(birnbaum(analyzer.guessingProbability, analyzer.actualDifficulty, theta) - sigma),
		}

		positiveConfidenceIntervalPoints = append(positiveConfidenceIntervalPoints, pointPositive)
		negativeConfidenceIntervalPoints = append(negativeConfidenceIntervalPoints, pointNegative)
	}

	return positiveConfidenceIntervalPoints, negativeConfidenceIntervalPoints
}

// * БЛОК ОТДЕЛЬНЫХ ФУНКЦИЙ

// Парсит csv-файл с результатами и теттами в соответствующий срез
func parseThetaResultPairs(filePath string) []thetaResultPair {
	file, err := os.Open(filePath)
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

// Вычисление минимальной и максимальной тетты в наборе
func minMaxTheta(pairs []thetaResultPair) (float64, float64) {
	min := 10.0
	max := -10.0

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

// Вычисление количества успешных выполнений задания в наборе
func goodAnswersAmount(pairs []thetaResultPair) int {
	amount := 0
	for _, value := range pairs {
		if value.result {
			amount++
		}
	}

	return amount
}

// Функция Бирнбаума
func birnbaum(c float64, delta float64, theta float64) float64 {
	exponent := math.Exp(1.71 * (theta - delta))
	return c + (1-c)*(exponent/(1+exponent))
}
