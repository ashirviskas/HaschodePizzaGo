package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
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

	valid := 0
	figures := generateFigures(minToppings, maxArea)
	slices := slicePizza(pizza, figures, int16(minToppings), int16(maxArea))
	fmt.Println(valid)
	writeSlicesToFile(output_f, slices)

	//fmt.Printf("wrote %d bytes\n", n4)

}

func slicePizza(pizza [][]int16, figures []coord, L, H int16) (slices []pizzaSlice) {
	scoreMap := make([][]int16, len(pizza))
	scoreMap_s := make([][][]int16, len(pizza))
	evaluations := make([][][]bool, len(pizza))
	scoreMap, scoreMap_s, evaluations = buildScoreMap(pizza, figures, true, L,
		coord{}, coord{}, scoreMap, scoreMap_s, evaluations)
	//scoreMap, scoreMap_s = buildScoreMap(pizza, figures, false, coord{0, 0}, coord{30, 30}, scoreMap, scoreMap_s,evalutionFn)
	start := time.Now()
	topMin := getTopMin(scoreMap_s, 20, true, coord{}, coord{}, make([]coordScore, 0), figures, 100000)
	valid := 0
	elapsed := time.Since(start)
	fmt.Printf("Calculating top 1 took %s\n", elapsed)
	i, j := 0, 0
	iterations := 0
	for len(topMin) > 0 {
		//fmt.Println(topMin)
		//break
		iterations++
		sliced := false
		fig := figures[topMin[0].fig]
		for tm, _ := range topMin {
			topM := topMin[len(topMin)-1-tm]
			//topM := topMin[tm]
			fig = figures[topM.fig]
			i, j = topM.place.row, topM.place.col
			s := pizzaSlice{coord{i, j}, coord{i + fig.row - 1, j + fig.col - 1}}
			valid_slice := sliceSlice(pizza, s, L, H, scoreMap_s, scoreMap)
			if valid_slice > 0 {
				valid += valid_slice
				sliced = true
				slices = append(slices, s)
				fmt.Println("Skipped", tm)
				break
			} else {
				//fmt.Println("REKT")
				scoreMap_s[i][j][topM.fig] = -1
				//scoreMap_s[i][j][topM.fig] = -1
			}
		}
		if sliced {
			//fmt.Println("Coordies", i, j)
			full_recalculate := false
			if iterations%100 == 0 {
				full_recalculate = true
			}
			scoreMap, scoreMap_s, evaluations = buildScoreMap(pizza, figures, full_recalculate, L,
				coord{i - 100, j - 100}, coord{i + 100 + fig.row, j + 100 + fig.col},
				scoreMap, scoreMap_s, evaluations)
			topMin = getTopMin(scoreMap_s, 150, true,
				coord{}, coord{},
				make([]coordScore, 0), figures, 5000)
			fmt.Println(valid)
			fmt.Println(iterations)
		} else {
			fmt.Println("Coordies", i, j)
			fmt.Println("Whoopsie")
			scoreMap, scoreMap_s, evaluations = buildScoreMap(pizza, figures, true, L,
				coord{i - 30, j - 30}, coord{i + 30 + fig.row, j + 30 + fig.col},
				scoreMap, scoreMap_s, evaluations)

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

func sliceSlice(pizza [][]int16, slice pizzaSlice, L, H int16, scoreMap_s [][][]int16, scoreMap [][]int16) (valid int) {
	mushroom, tomato, total := evalSlice(pizza, slice)
	valid = 0
	if mushroom >= L && tomato >= L && total <= H {
		for k := slice.start.row; k <= slice.end.row; k++ {
			for t := slice.start.col; t <= slice.end.col; t++ {
				pizza[k][t] = 0
				valid++
				scoreMap[k][t] = -1
				for l, _ := range scoreMap_s[k][t] {
					scoreMap_s[k][t][l] = -1
				}
			}
		}
	} else {

	}
	return
}

func getTopMin(scoreMap_s [][][]int16, maxTake uint16, full_recalculate bool,
	start_c, end_c coord, topMin_o []coordScore, figures []coord, max_iter int) (topMin []coordScore) {
	if !full_recalculate {
		topMin = topMin_o
	}
	maxTopMinScore := int16(16000)
	iters := 0
	for i, row := range scoreMap_s {
		if iters > max_iter {
			return
		}
		if !full_recalculate && (i < start_c.row || i > end_c.row) {
			continue
		}

		for j, col := range row {
			if !full_recalculate && (j < start_c.col || j > end_c.col) {
				continue
			}
			if col[len(col)-1] <= 0 {
				continue
			}
			smallest_sc := coordScore{coord{i, j}, maxTopMinScore, -1, 1}
			for f, sc := range col {
				if f == len(col)-1 {
					break
				}
				fig_area := int16(figures[f].row * figures[f].col)
				//fmt.Println("LOL", fig_area)
				if sc > 0 {
					if (sc<<4)/fig_area < smallest_sc.score {
						smallest_sc.score = (sc << 4) / fig_area
						smallest_sc.fig = int16(f)
						smallest_sc.area = fig_area
					}
				}
			}
			//fmt.Println(smallest_sc.score)
			if smallest_sc.score < maxTopMinScore {
				if uint16(len(topMin)) < maxTake {
					topMin = append(topMin, smallest_sc)
					normalSort(topMin)
					maxTopMinScore = topMin[0].score
				} else {
					if smallest_sc.score < maxTopMinScore {
						for tp, cS := range topMin {
							//fmt.Println(cS)
							if tp == len(topMin)-1 {
								topMin = append(topMin[1:len(topMin)], smallest_sc)
								//topMin[len(topMin)-1] = smallest_sc
								break
							}
							//} else {
							if smallest_sc.score < cS.score && tp != len(topMin)-1 {
								//print(smallest_sc.score)
								continue
							} else {
								topMin_t := append(topMin[1:tp+1], smallest_sc)
								if tp < len(topMin)-1 {
									topMin = append(topMin_t, topMin[tp:len(topMin)-1]...)
								} else {
									topMin = topMin_t
								}
								break
							}

						}
					}

				}
			}

		}
	}
	return
}

func normalSort(items []coordScore) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].score > items[j].score
	})
}

func evaluationFn(pizza [][]int16, slice pizzaSlice, L int16) (valid bool, total int16) {
	mushroom, tomato, total := evalSlice(pizza, slice)
	if mushroom >= L && tomato >= L {
		return true, total
	}
	return false, total
}

func buildScoreMap(pizza [][]int16, figures []coord, full_recalculate bool, L int16,
	start_c, end_c coord, scoreBoi_i [][]int16, scoreBoi_si [][][]int16, evaluations_i [][][]bool,
) (scoreBoi [][]int16, scoreBoi_s [][][]int16, evaluations [][][]bool) {
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
			scoreBoi[i][j] = 0
			if cel == 0 {
				continue
			}
			for f, fig := range figures {
				s := pizzaSlice{coord{i, j}, coord{i + fig.row - 1, j + fig.col - 1}}
				ev, t := evaluationFn(pizza, s, L)
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
				if scoreBoi_s[i][j][len(figures)] == -1 {
					break
				}
				if scoreBoi_s[i][j][f] == -1 {
					continue
				}
				scoreBoi_s[i][j][f] = 0
				scoreBoi_s[i][j][len(figures)] = 0
				if ev {
					for r := i; r <= i+fig.row-1; r++ {
						for c := j; c <= j+fig.col-1; c++ {
							scoreBoi_s[i][j][f] += scoreBoi[r][c]
							scoreBoi_s[i][j][len(figures)] += scoreBoi[r][c]
						}
					}
				}
			}
		}
	}
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
	for i := s.start.row; i <= s.end.row; i++ {
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
			total++
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
