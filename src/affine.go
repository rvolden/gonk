package main

import (
    "fmt"
    "flag"
    "bufio"
    "os"
    "log"
    "time"
    "strings"
)

// Command line arguments
var seqA = flag.String("a", "", "First sequence to align")
var seqB = flag.String("b", "", "Second sequence to align")
var diag = flag.Bool("d", false, "Use to exclude the main diagonal")
var open = flag.Int("o", 25, "Gap open penalty")
var ext = flag.Int("e", 1, "Gap extend penalty")
var mat = flag.Bool("m", false, "Output entire matrix to stdout")
var out = flag.String("out", "SW_PARSE.txt", "Where to output file")

type fasta struct {
	header, seq string
}

func (fa *fasta) add_seq(sequence string) {
	fa.seq += sequence
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func fasta_reader(seqFile *string) (seqs []fasta) {
    // open up the file
	file, err := os.Open(*seqFile)
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
        if line == "" {
            continue
        }
        // look for >
		if line[0] == 62 {
            // make a slice entry with no seq
			var entry = fasta{string(line[1:]), ""}
			seqs = append(seqs, entry)
		} else {
            // if on a line without >, add the sequence
            // to the last header
			seqs[len(seqs)-1].add_seq(line)
		}
	}
	return seqs
}

func makeMatrix(n, m int) [][]int {
    // makes an n by m matrix
    matrix := make([][]int, n + 1)
    for i := 0; i < n + 1; i++ {
        matrix[i] = make([]int, m + 1)
    }
    return matrix
}

func maxScore(scores [3]int) int {
    max := 0
    for _, score := range scores {
        if score > max {
            max = score
        }
    }
    return max
}

func populate(matches, gapA, gapB [][]int, seqA, seqB string) [][]int {
    // Fills in score matrix
    matrices := [3][][]int{matches, gapA, gapB}

    // map to make indexing score matrix easier
    baseInt := map[int]int{
        // A, C, G, T
        65:0, 67:1, 71:2, 84:3,
    }

    // match and mismatch scores
    scores := [][]int{
        {5, -4, -4, -4},
        {-4, 5, -4, -4},
        {-4, -4, 5, -4},
        {-4, -4, -4, 5},
    }

    // goes through and gets scores
    for i := 1; i < len(matches); i++ {
        for j := 1; j < len(matches[0]); j++ {
            // exclude the diagonal depending on the flag
            if i == j && *diag {
                matches[i][j] = 0
            } else {
                maxes := [3]int{}
                indA, indB := int(seqA[i - 1]), int(seqB[j - 1])
                score := scores[baseInt[indA]][baseInt[indB]]

                for k, matrix := range matrices {
                    // potential scores to put into matrix
                    potential := [3]int{}

                    if k == 0 { // match/mismatch matrix
                        potential[0] = matches[i-1][j-1] + score
                        potential[1] = gapA[i-1][j-1] + score
                        potential[2] = gapB[i-1][j-1] + score
                    } else if k == 1 { // gap in seqA
                        potential[0] = -*open - *ext + matches[i][j-1]
                        potential[1] = -*ext + gapA[i][j-1]
                        potential[2] = -*open - *ext + gapB[i][j-1]
                    } else if k == 2 { // gap in seqB
                        potential[0] = -*open - *ext + matches[i-1][j]
                        potential[1] = -*open - *ext + gapA[i-1][j]
                        potential[2] = -*ext + gapB[i-1][j]
                    }
                    maxes[k] = maxScore(potential)
                    matrix[i][j] = maxes[k]
                }
                toAdd := maxScore(maxes)
                matches[i][j] = toAdd
            }
        }
    }
    return matches
}

func sums(matches [][]int) []int {
    // Gives summed scores along diagonals
    sums := make([]int, len(matches[0]))
    for i := 0; i < len(matches); i++ {
        for j := i; j < len(matches[0]); j++ {
            sums[j-i] += matches[i][j]
        }
    }
    return sums
}

func printMatrix(matchMatrix [][]int) {
    // Prints out the entire score matrix to stdout
    for f := 0; f < len(matchMatrix); f++ {
        row := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(matchMatrix[f])), " "),"[]")
        fmt.Println(row)
    }
}

func writeScores(sums []int) {
    // Writes the summed scores to the outfile
    outFile, err := os.Create(*out)
    check(err)
    outWrite := bufio.NewWriter(outFile)
    for i, score := range sums {
        _, err := fmt.Fprintf(outWrite, "%v:%v\n", i, score)
        check(err)
    }
    outWrite.Flush()
}

func main() {
    // start timer and parse flags
    start := time.Now()
    flag.Parse()
    if *seqA == "" || *seqB == "" {
        log.Fatal("Please enter sequences\n")
    }

    // read fasta files and get both seqs
    first := fasta_reader(seqA)[0].seq
    second := fasta_reader(seqB)[0].seq
    n, m := len(first), len(second)

    // construct the alignment matrix (2D slice) and populate
    matches, gapA, gapB := makeMatrix(n, m), makeMatrix(n, m), makeMatrix(n, m)
    matchMatrix := populate(matches, gapA, gapB, first, second)
    sumList := sums(matchMatrix)

    // print out the matrix
    if *mat {
        printMatrix(matches)
    }

    // write scores
    writeScores(sumList)

    // parameters to stderr for debugging
    fmt.Fprintf(os.Stderr, "Running parameters:\nseqA: %v\n", *seqA)
    fmt.Fprintf(os.Stderr, "seqB: %v\n", *seqB)
    fmt.Fprintf(os.Stderr, "diag: %v\n", *diag)
    fmt.Fprintf(os.Stderr, "matrix: %v\n", *mat)
    fmt.Fprintf(os.Stderr, "gap open: %v\n", *open)
    fmt.Fprintf(os.Stderr, "gap extend: %v\n", *ext)
    fmt.Fprintf(os.Stderr, "out: %v\n", *out)

    // stop timer and print elapsed time to stderr
    stop := time.Now()
    elapsed := stop.Sub(start)
    fmt.Fprintf(os.Stderr, "Took %v to run.\n", elapsed)
}
