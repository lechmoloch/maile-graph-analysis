package TaskAnalysis

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
)

// * ВСПОМОГАТЕЛЬНЫЕ СТРУКТУРЫ

// Точка графика
type graphPoint struct {
	Theta     float64
	Frequence float64
}

// Пара результат - уровень подготовленности
type thetaResultPair struct {
	Theta  float64
	Result bool
}

// Вычисленные данные графического метода
type GraphicalMethodData struct {
	Verdict bool
	Points  [][4]graphPoint
}

// Вычисленные данные метода на основе гипотез
type HypothesisMethodData struct {
	Verdict                        string
	CorrectTaskLikelihoodRatio     float64
	IndifferentTaskLikelihoodRatio float64
	IncorrectTaskLikelihoodRatio   float64
}

// Собрание данных всех методов
type AllMethodsData struct {
	GraphicalMethodData  GraphicalMethodData
	HypothesisMethodData HypothesisMethodData
}

type TaskAnalyzer struct {
	actualDifficulty    float64
	guessingProbability float64
	resultPairs         []thetaResultPair
}

// * ПУБЛИЧНЫЕ МЕТОДЫ

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

// Парсинг вычисленных данных в JSON
func (analyzer TaskAnalyzer) WriteAllMethodsDataToJSON(outputFilePath string) {
	pointsActual, pointsBirnbaum, pointsPositiveInterval, pointsNegativeInterval := analyzer.calculateAllGraphPoints()

	var allGraphPoints [][4]graphPoint
	for i := 0; i < len(pointsActual); i++ {
		var currentGraphPoints [4]graphPoint = [4]graphPoint{
			pointsActual[i],
			pointsBirnbaum[i],
			pointsPositiveInterval[i],
			pointsNegativeInterval[i],
		}

		allGraphPoints = append(allGraphPoints, currentGraphPoints)
	}

	var GraphicalMethodData GraphicalMethodData = GraphicalMethodData{
		Verdict: analyzer.makeVerdictGraphicalMethod(),
		Points:  allGraphPoints,
	}

	var HypothesisMethodData HypothesisMethodData = HypothesisMethodData{
		Verdict:                        analyzer.makeHypothesisMethodVerdict(),
		CorrectTaskLikelihoodRatio:     analyzer.calculateCorrectTaskLikelihoodRatio(),
		IndifferentTaskLikelihoodRatio: analyzer.calculateIndifferentTaskLikelihoodRatio(),
		IncorrectTaskLikelihoodRatio:   analyzer.calculateIncorrectTaskLikelihoodRatio(),
	}

	var AllMethodsData AllMethodsData = AllMethodsData{
		GraphicalMethodData:  GraphicalMethodData,
		HypothesisMethodData: HypothesisMethodData,
	}

	file, _ := json.MarshalIndent(AllMethodsData, "", " ")
	_ = os.WriteFile(outputFilePath, file, 0644)
}

// Производит csv-файл с точками для построения графика
func (analyzer TaskAnalyzer) WriteGraphDataToCSV(outputFilePath string) {
	file, err := os.Create(outputFilePath)
	if err != nil {
		panic(err)
	}

	pointsActual, pointsBirnbaum, pointsPositiveInterval, pointsNegativeInterval := analyzer.calculateAllGraphPoints()

	writer := csv.NewWriter(file)
	for i := 0; i < len(pointsActual); i++ {
		var record [8]string = [8]string{
			strconv.FormatFloat(pointsActual[i].Theta, 'f', -1, 64),
			strconv.FormatFloat(pointsActual[i].Frequence, 'f', -1, 64),
			strconv.FormatFloat(pointsBirnbaum[i].Theta, 'f', -1, 64),
			strconv.FormatFloat(pointsBirnbaum[i].Frequence, 'f', -1, 64),
			strconv.FormatFloat(pointsPositiveInterval[i].Theta, 'f', -1, 64),
			strconv.FormatFloat(pointsPositiveInterval[i].Frequence, 'f', -1, 64),
			strconv.FormatFloat(pointsNegativeInterval[i].Theta, 'f', -1, 64),
			strconv.FormatFloat(pointsNegativeInterval[i].Frequence, 'f', -1, 64),
		}
		if err := writer.Write(record[:]); err != nil {
			panic(err)
		}
	}

	writer.Flush()
}

// * МЕТОДЫ ДЛЯ ГРАФИЧЕСКОГО АНАЛИЗА

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
			if currentPair.Theta == currentPocket {
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
		Theta := currentPocketPairs[0].Theta
		Frequence := float64(goodAnswersAmount(currentPocketPairs)) / float64(len(currentPocketPairs))

		var currentPoint graphPoint = graphPoint{Theta, Frequence}
		actualPoints = append(actualPoints, currentPoint)
	}

	return actualPoints
}

// Вычисление точек предполагаемых результатов
func (analyzer TaskAnalyzer) calculateBirnbaumPoints(pocketPairs [][]thetaResultPair) []graphPoint {
	var birnbaumPoints []graphPoint
	for _, currentPocket := range pocketPairs {
		var currentPoint graphPoint
		currentPoint.Theta = currentPocket[0].Theta
		currentPoint.Frequence = birnbaum(
			analyzer.guessingProbability,
			analyzer.actualDifficulty,
			currentPocket[0].Theta,
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
		Frequence := float64(goodAnswersAmount(currentPocketPairs)) / float64(len(currentPocketPairs))
		sigma := math.Sqrt(studentsAmount * Frequence * (1 - Frequence))

		Theta := currentPocketPairs[0].Theta
		var pointPositive, pointNegative graphPoint

		pointPositive = graphPoint{
			Theta,
			(birnbaum(analyzer.guessingProbability, analyzer.actualDifficulty, Theta) + sigma),
		}
		pointNegative = graphPoint{
			Theta,
			(birnbaum(analyzer.guessingProbability, analyzer.actualDifficulty, Theta) - sigma),
		}

		positiveConfidenceIntervalPoints = append(positiveConfidenceIntervalPoints, pointPositive)
		negativeConfidenceIntervalPoints = append(negativeConfidenceIntervalPoints, pointNegative)
	}

	return positiveConfidenceIntervalPoints, negativeConfidenceIntervalPoints
}

// Вычисление всех точек разом
func (analyzer TaskAnalyzer) calculateAllGraphPoints() ([]graphPoint, []graphPoint, []graphPoint, []graphPoint) {
	pocketPairs := analyzer.getResultPocketPairs()
	pointsActual := analyzer.calculateActualPoints(pocketPairs)
	pointsBirnbaum := analyzer.calculateBirnbaumPoints(pocketPairs)
	pointsPositiveInterval, pointsNegativeInterval := analyzer.calculateConfidenceIntervalsPoints(pocketPairs)

	return pointsActual, pointsBirnbaum, pointsPositiveInterval, pointsNegativeInterval
}

// Вычисление вердикта по графическому методу
func (analyzer TaskAnalyzer) makeVerdictGraphicalMethod() bool {
	pocketPairs := analyzer.getResultPocketPairs()
	pointsActual := analyzer.calculateActualPoints(pocketPairs)
	pointsPositiveInterval, pointsNegativeInterval := analyzer.calculateConfidenceIntervalsPoints(pocketPairs)

	Verdict := true
	for i := 0; i < len(pointsActual); i++ {
		pocketIsNotEmpty := len(pocketPairs[i]) > 0
		fmt.Print(strconv.Itoa(i) + ": ")
		fmt.Println(len(pocketPairs[i]))
		if (pointsActual[i].Frequence > pointsPositiveInterval[i].Frequence || pointsActual[i].Frequence < pointsNegativeInterval[i].Frequence) && pocketIsNotEmpty {
			Verdict = false
			break
		}
	}

	return Verdict
}

// * МЕТОДЫ ДЛЯ АНАЛИЗА НА ОСНОВЕ ГИПОТЕЗ

// Отношение правдоподобия для гипотезы о корректности задания
func (analyzer TaskAnalyzer) calculateCorrectTaskLikelihoodRatio() float64 {
	sum := 0.0
	for _, value := range analyzer.resultPairs {
		if value.Result {
			sum += math.Log(birnbaum(analyzer.guessingProbability, analyzer.actualDifficulty, value.Theta))
		} else {
			sum += math.Log(1 - birnbaum(analyzer.guessingProbability, analyzer.actualDifficulty, value.Theta))
		}
	}

	return sum
}

// Отношение правдоподобия для гипотезы об индефферентности задания
func (analyzer TaskAnalyzer) calculateIndifferentTaskLikelihoodRatio() float64 {
	return -1.0 * float64(len(analyzer.resultPairs)) * math.Ln2
}

// Отношение правдоподобия для гипотезы о некорректности задания
func (analyzer TaskAnalyzer) calculateIncorrectTaskLikelihoodRatio() float64 {
	sum := 0.0
	for _, value := range analyzer.resultPairs {
		exponent := math.Exp(-1 * 1.71 * (value.Theta - analyzer.actualDifficulty))
		localBirnbaum := analyzer.guessingProbability + (1-analyzer.guessingProbability)*(exponent/(1+exponent))
		if value.Result {
			sum += math.Log(localBirnbaum)
		} else {
			sum += math.Log(1 - localBirnbaum)
		}
	}

	return sum
}

func (analyzer TaskAnalyzer) makeHypothesisMethodVerdict() string {
	var likelihoodRatios []float64
	likelihoodRatios = append(likelihoodRatios, analyzer.calculateCorrectTaskLikelihoodRatio())
	likelihoodRatios = append(likelihoodRatios, analyzer.calculateIndifferentTaskLikelihoodRatio())
	likelihoodRatios = append(likelihoodRatios, analyzer.calculateIncorrectTaskLikelihoodRatio())

	sort.Float64s(likelihoodRatios)
	var Verdict string
	switch likelihoodRatios[len(likelihoodRatios)-1] {

	case analyzer.calculateCorrectTaskLikelihoodRatio():
		Verdict = "Correct " + strconv.FormatFloat(analyzer.calculateCorrectTaskLikelihoodRatio(), 'f', -1, 64)

	case analyzer.calculateIndifferentTaskLikelihoodRatio():
		Verdict = "Indifferent " + strconv.FormatFloat(analyzer.calculateIndifferentTaskLikelihoodRatio(), 'f', -1, 64)

	case analyzer.calculateIncorrectTaskLikelihoodRatio():
		Verdict = "Incorrect " + strconv.FormatFloat(analyzer.calculateIncorrectTaskLikelihoodRatio(), 'f', -1, 64)

	}

	return Verdict
}

// * БЛОК ВСПОМОГАТЕЛЬНЫХ ФУНКЦИЙ ДЛЯ ГРАФИЧЕСКОГО АНАЛИЗА
// TODO Вынести в методы

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
		pair.Result = resultValue != 0
		pair.Theta = float64(thetaValue)

		pairs = append(pairs, pair)
	}

	return pairs
}

// Вычисление минимальной и максимальной тетты в наборе
func minMaxTheta(pairs []thetaResultPair) (float64, float64) {
	min := 10.0
	max := -10.0

	for _, value := range pairs {
		if value.Theta < min {
			min = value.Theta
		}

		if value.Theta > max {
			max = value.Theta
		}
	}

	return min, max
}

// Вычисление количества успешных выполнений задания в наборе
func goodAnswersAmount(pairs []thetaResultPair) int {
	amount := 0
	for _, value := range pairs {
		if value.Result {
			amount++
		}
	}

	return amount
}

// Функция Бирнбаума
func birnbaum(c float64, delta float64, Theta float64) float64 {
	exponent := math.Exp(1.71 * (Theta - delta))
	return c + (1-c)*(exponent/(1+exponent))
}
