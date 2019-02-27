package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	//input_f := "inputs/small.in"
	//output_f := "outputs/small.out"
	//input_f := "inputs/example.in"
	//output_f := "outputs/example.out"
	input_f := "inputs/medium.in"
	output_f := "outputs/medium.out"
	file, err := os.Open(input_f)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var rows, cols, minToppings, maxArea int

	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		rows, err = strconv.Atoi(fields[0])
		cols, err = strconv.Atoi(fields[1])
		minToppings, err = strconv.Atoi(fields[2])
		maxArea, err = strconv.Atoi(fields[3])

		fmt.Println()
	}

	pizza_str := make([][]string, rows)

	var i = 0
	for scanner.Scan() {
		col := strings.Split(scanner.Text(), "")
		pizza_str[i] = col
		i = i + 1
	}
	pizza := make([][]int16, len(pizza_str))

	for i := range pizza {
		pizza[i] = make([]int16, len(pizza_str[0]))
	}

	{
		for i, row := range pizza_str {
			for j, col := range row {
				switch col {
				case "M":
					pizza[i][j] = 1
				default:
					pizza[i][j] = 2
				}
			}
		}
	}

	fmt.Println(rows, cols, minToppings, maxArea)

	evalutionFn := getEvaluationFn(pizza, int16(minToppings), int16(maxArea))
	valid := 0
	figures := generateFigures(minToppings, maxArea)
	slices := slicePizza(pizza, figures, int16(minToppings), int16(maxArea), evalutionFn)

	//for i := 0; i < len(pizza); i++ {
	//	for j := 0; j < len(pizza[0]); j++ {
	//		for _, fig := range figures {
	//			if i+fig.row < len(pizza) && j+fig.col < len(pizza[0]) {
	//				s := pizzaSlice{coord{i, j}, coord{i + fig.row-1, j + fig.col-1}}
	//				ev, t := evalutionFn(s)
	//				if t > 14 {
	//					fmt.Println(t)
	//				}
	//				if ev == true {
	//					for k := s.start.row; k <= s.end.row; k ++ {
	//						for t := s.start.col; t <= s.end.col; t++ {
	//							pizza[k][t] = 0
	//							valid++
	//						}
	//					}
	//					slices = append(slices, s)
	//				}
	//			}
	//		}
	//
	//	}
	//}
	fmt.Println(valid)
	writeSlicesToFile(output_f, slices)

	//fmt.Printf("wrote %d bytes\n", n4)

}

func slicePizza(pizza [][]int16, figures []coord, L, H int16, evalutionFn func(slice pizzaSlice) (ok bool, score int16)) (slices []pizzaSlice) {
	scoreMap := make([][]int16, len(pizza))
	scoreMap_s := make([][][]int16, len(pizza))
	evaluations := make([][][]bool, len(pizza))
	scoreMap, scoreMap_s, evaluations = buildScoreMap(pizza, figures, true,
		coord{}, coord{}, scoreMap, scoreMap_s, evaluations, evalutionFn)
	//scoreMap, scoreMap_s = buildScoreMap(pizza, figures, false, coord{0, 0}, coord{30, 30}, scoreMap, scoreMap_s,evalutionFn)
	start := time.Now()
	topMin := getTopMin(scoreMap_s, 100, true, coord{}, coord{}, make([]coordScore, 0), figures, 100000)
	valid := 0
	elapsed := time.Since(start)
	fmt.Printf("Calculating top 1 took %s\n", elapsed)
	i, j := 0, 0
	for len(topMin) > 0 {
		sliced := false
		fig := figures[topMin[0].fig]
		for tm, _ := range topMin{
			topM := topMin[len(topMin) -1 - tm]
			//topM := topMin[tm]
			fig = figures[topM.fig]
			i, j = topM.place.row, topM.place.col
			s := pizzaSlice{coord{i, j}, coord{i + fig.row-1, j + fig.col-1}}
			valid_slice := sliceSlice(pizza, s, L, H)
			if valid_slice > 0{
				valid += valid_slice
				sliced = true
				slices = append(slices, s)
				fmt.Println("Skipped",tm)
				break
			} else{
				scoreMap_s[i][j][topM.fig] = -1
				//scoreMap_s[i][j][topM.fig] = -1
			}
		}
		if sliced {
			//fmt.Println("Coordies", i, j)
			scoreMap, scoreMap_s, evaluations = buildScoreMap(pizza, figures, false,
				coord{i - 24, j - 24}, coord{i + 24 + fig.row, j + 24 + fig.col},
				scoreMap, scoreMap_s, evaluations,evalutionFn)
			topMin = getTopMin(scoreMap_s, 50, true,
				coord{}, coord{},
				make([]coordScore, 0), figures, 3000)
			fmt.Println(valid)
		} else{
			fmt.Println("Coordies", i, j)
			fmt.Println("Whoopsie")
			//scoreMap, scoreMap_s = buildScoreMap(pizza, figures, false,
			//	coord{fig.row - 30, fig.col - 30}, coord{fig.row + 30, fig.col + 30},
			//	scoreMap, scoreMap_s, evalutionFn)

			topMin = getTopMin(scoreMap_s, 3000, true,
				coord{}, coord{},
				make([]coordScore, 0), figures, 50000)
			fmt.Println(valid)
		}

	}

	fmt.Println(topMin)
	fmt.Println(len(topMin))

	return
}

func sliceSlice(pizza [][]int16, slice pizzaSlice, L, H int16) (valid int){
	mushroom, tomato, total := evalSlice(pizza, slice)
	if mushroom >= L && tomato >= L && total <= H {
		for k := slice.start.row; k <= slice.end.row; k ++ {
			for t := slice.start.col; t <= slice.end.col; t++ {
				pizza[k][t] = 0
				valid++
			}
		}
	}
	return
}


func getTopMin(scoreMap_s [][][]int16, maxTake uint16, full_recalculate bool,
	start_c, end_c coord, topMin_o []coordScore, figures []coord, max_iter int) (topMin []coordScore) {
	if !full_recalculate {
		topMin = topMin_o
	}
	maxTopMinScore := int16(9999)
	iters := 0
	for i, row := range scoreMap_s {
		if iters > max_iter{
			return
		}
		if !full_recalculate && (i < start_c.row || i > end_c.row) {
			continue
		}

		for j, col := range row {
			if !full_recalculate && (j < start_c.col || j > end_c.col) {
				continue
			}
			if col[len(col)-1] == 0 {
				continue
			}
			smallest_sc := coordScore{coord{i, j}, 9999, 0, 2}
			for f, sc := range col {
				if f == len(col) -1{
					break
				}
				fig_area := int16(figures[f].row * figures[f].col)
				//fmt.Println("LOL", fig_area)
				if sc > 0 {
					if (sc << 4) / fig_area < (smallest_sc.score << 4) / smallest_sc.area {
						smallest_sc.score = sc
						smallest_sc.fig = int16(f)
						smallest_sc.area = fig_area
					}
				}
			}
			//fmt.Println(smallest_sc.score)
			if smallest_sc.score < 9999 {
				if uint16(len(topMin)) < maxTake {
					topMin = append(topMin, smallest_sc)
					insertionsort(topMin)
					maxTopMinScore = (topMin[0].score << 4) / topMin[0].area
				} else {
					if (smallest_sc.score << 4) / smallest_sc.area < maxTopMinScore {
						for tp, cS := range topMin {
							//fmt.Println(cS)
								if tp == len(topMin)-1 {
									topMin = append(topMin[1:len(topMin)], smallest_sc)
									//topMin[len(topMin)-1] = smallest_sc
									break
								}
							//} else {
								if (smallest_sc.score << 4) / smallest_sc.area < (cS.score << 4) / cS.area && tp != len(topMin)-1{
									//print(smallest_sc.score)
									continue
								} else {
									topMin_t := append(topMin[1:tp+1], smallest_sc)
									//fmt.Println(topMin_t)
									//fmt.Println(len(topMin_t))
									//fmt.Println(topMin)
									if tp < len(topMin) -1 {
										topMin = append(topMin_t, topMin[tp:len(topMin)-1]...)
									} else{
										topMin = topMin_t
									}
									break
								}

							//}
						}
					}

				}
			}

		}
	}
	return
}

func getEvaluationFn(pizza [][]int16, L, H int16) (func(slice pizzaSlice) (ok bool, score int16)) {
	return func(slice pizzaSlice) (ok bool, score int16) {
		mushroom, tomato, total := evalSlice(pizza, slice)
		if mushroom >= L && tomato >= L && total <= H {
			return true, total
		}
		return false, total
	}
}

func insertionsort(items []coordScore) {
	var n = len(items)
	for i := 1; i < n; i++ {
		j := i
		for j > 0 {
			if items[j-1].score-items[j-1].area < items[j].score-items[j].area {
				items[j-1], items[j] = items[j], items[j-1]
			}
			j = j - 1
		}
	}
}

func buildScoreMap(pizza [][]int16, figures []coord, full_recalculate bool,
	start_c, end_c coord, scoreBoi_i [][]int16, scoreBoi_si [][][]int16, evaluations_i [][][]bool,
	evaluationFn func(slice pizzaSlice) (ok bool, score int16)) (scoreBoi [][]int16, scoreBoi_s [][][]int16, evaluations [][][]bool) {
	//start := time.Now()

	figure_count := len(figures)
	evaluations = evaluations_i
	if full_recalculate {
		evaluations = make([][][]bool, len(pizza))
		for i := range evaluations {
			evaluations[i] = make([][]bool, len(pizza[0]))
			for j := range evaluations[i] {
				evaluations[i][j] = make([]bool, figure_count)
			}
		}
		scoreBoi = make([][]int16, len(pizza))
		scoreBoi_s = make([][][]int16, len(pizza))
		for i := range scoreBoi {
			scoreBoi[i] = make([]int16, len(pizza[0]))

		}

		for i := range scoreBoi_s {
			scoreBoi_s[i] = make([][]int16, len(pizza[0]))
			evaluations[i] = make([][]bool, len(pizza[0]))
			for j := range scoreBoi[i] {
				scoreBoi_s[i][j] = make([]int16, figure_count+1)
				evaluations[i][j] = make([]bool, figure_count)
			}
		}
	} else {
		scoreBoi = scoreBoi_i
		scoreBoi_s = scoreBoi_si
	}


	//start := time.Now()
	for i, row := range pizza {
		if !full_recalculate && (i < start_c.row || i > end_c.row) {
			continue
		}
		for j, cel := range row {
			if !full_recalculate && (j < start_c.col || j > end_c.col) {
				continue
			}

			if cel == 0 {
				scoreBoi[i][j] = 0
				continue
			}
			for f, fig := range figures {
				s := pizzaSlice{coord{i, j}, coord{i + fig.row - 1, j + fig.col - 1}}
				ev, t := evaluationFn(s)
				evaluations[i][j][f] = ev
				if ev {
					scoreBoi[i][j] += t
				}
			}
		}
	}

	for i, row := range scoreBoi {
		if !full_recalculate && (i < start_c.row || i > end_c.row) {
			continue
		}
		for j, cel := range row {
			if !full_recalculate && (j < start_c.col || j > end_c.col) {
				continue
			}
			if cel == 0 {
				continue
			}
			for f, fig := range figures {
				ev := evaluations[i][j][f]
				if scoreBoi_s[i][j][len(figures)] == -1{
					break
				}
				if scoreBoi_s[i][j][f] == -1{
					continue
				}
				if !full_recalculate{
					scoreBoi_s[i][j][f] = 0
					scoreBoi_s[i][j][len(figures)] = 0
				}
				if ev {
					for r := i; r <= i+fig.row-1; r ++ {
						for c := j; c <= j+fig.col-1; c++ {
							scoreBoi_s[i][j][f] += scoreBoi[r][c]
							scoreBoi_s[i][j][len(figures)] += scoreBoi[r][c]
						}
					}
				}
			}
		}
	}
	//elapsed := time.Since(start)
	//fmt.Printf("Calculating scores took %s\n", elapsed)
	//fmt.Print(scoreBoi[0][0], "\n")
	//fmt.Print(scoreBoi[10][10], "\n")
	//fmt.Print(scoreBoi[20][20], "\n")
	return
}

func writeSlicesToFile(filename string, slices []pizzaSlice) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	w := bufio.NewWriter(f)
	var slice_n int
	slice_n = len(slices)
	slice_n_str := strconv.Itoa(slice_n)
	_, err = w.WriteString(slice_n_str + "\n")
	for _, slc := range slices {
		w.WriteString(strconv.Itoa(slc.start.row) + " " + strconv.Itoa(slc.start.col) + " " +
			strconv.Itoa(slc.end.row) + " " + strconv.Itoa(slc.end.col) + "\n")
	}
	w.Flush()
}

func evalSlice(pizza [][]int16, s pizzaSlice) (mushroom, tomato, total int16) {
	if s.end.row >= len(pizza) || s.end.col >= len(pizza[0]) {
		return 0, 0, 0
	}
	for i := s.start.row; i <= s.end.row; i ++ {
		for j := s.start.col; j <= s.end.col; j++ {

			if pizza[i][j] == 1 {
				mushroom++
			}

			if pizza[i][j] == 2 {
				tomato++
			}
			if pizza[i][j] == 0 {
				tomato = 0
				mushroom = 0
				total = 0
				return
			}
			total ++
		}

	}
	return
}

func generateFigures(minThings, maxSize int) (figures []coord) {
	for i := 1; i < maxSize/2+1; i++ {
		for j := maxSize; j > i-1; j-- {
			if maxSize >= i*j && i*j >= minThings*2 {
				figures = append(figures, coord{j, i})
				if i != j {
					figures = append(figures, coord{i, j})
				}
			}
		}
	}
	fmt.Println(figures)
	return
}

type coord struct {
	row, col int
}

type pizzaSlice struct {
	start, end coord
}
type coordScore struct {
	place            coord
	score, fig, area int16
}
