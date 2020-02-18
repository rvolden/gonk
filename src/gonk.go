package main

import (
    "bufio"
    "flag"
    "fmt"
    "io"
    "log"
    "os"
    str "strings"
    "time"
)

// Command line arguments
var seqA = flag.String("a", "", "First sequence to align")
var seqB = flag.String("b", "", "Second sequence to align")
var diag = flag.Bool("d", false, "Use to exclude the main diagonal")
var penalty = flag.Int("p", 25, "Gap open and extend penalty")
var mat = flag.Bool("m", false, "Output entire matrix to stdout")
var out = flag.String("o", "SW_PARSE.txt", "Where to output file")

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

func readFasta(seqFile *string) (seqs []fasta) {
    // open up the file
    file, err := os.Open(*seqFile)
    check(err)
    defer file.Close()

    reader := bufio.NewReader(file)
    for {
        line, e := reader.ReadString('\n')
        if e == io.EOF {
            break
        }
        line = str.TrimSuffix(line, "\n")
        if line == "" {
            continue
        }
        // look for >
        if line[0] == '>' {
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
    matrix := make([][]int, n+1)
    for i := 0; i < n+1; i++ {
        matrix[i] = make([]int, m+1)
    }
    return matrix
}

func maxScore(scores [4]int) int {
    // gives the max value from an array
    max := 0
    for _, score := range scores {
        if score > max {
            max = score
        }
    }
    return max
}

func populate(matches [][]int, seqA, seqB string) [][]int {
    // Fills in score matrix
    // map to make indexing score matrix easier
    baseInt := map[int]int{
        'A': 0, 'C': 1, 'G': 2, 'T': 3, 'N': 4,
    }

    // match and mismatch scores
    scores := [][]int{
        {5, -4, -4, -4, 0},
        {-4, 5, -4, -4, 0},
        {-4, -4, 5, -4, 0},
        {-4, -4, -4, 5, 0},
        {0, 0, 0, 0, 0},
    }

    // goes through and gets scores
    for i := 1; i < len(matches); i++ {
        for j := 1; j < len(matches[0]); j++ {
            // exclude the diagonal depending on the flag
            if i == j && *diag {
                matches[i][j] = 0
            } else {
                // potential scores to put into matrix
                potential := [4]int{}
                indA, indB := int(seqA[i-1]), int(seqB[j-1])
                // either a match or mismatch
                score := scores[baseInt[indA]][baseInt[indB]]
                potential[0] = matches[i-1][j-1] + score  // back
                potential[1] = matches[i-1][j] - *penalty // left
                potential[2] = matches[i][j-1] - *penalty // above
                max := maxScore(potential)
                matches[i][j] = max
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
        row := str.Trim(str.Join(str.Fields(fmt.Sprint(matchMatrix[f])), " "), "[]")
        fmt.Println(row)
    }
}

func checkOutFile() (fp *string) {
    // Checks if the input is a directory
    // Otherwise just return the same file
    if fileInfo, err := os.Stat(*out); err == nil {
        if fileInfo.Mode().IsDir() {
            // add on the filename if given a directory
            *out += "/SW_PARSE.txt"
            return out
        }
    } else if os.IsNotExist(err) {
        return out
    }
    return out
}

func writeScores(sums []int) {
    // Writes the summed scores to the outfile
    out = checkOutFile()
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
    first := readFasta(seqA)[0].seq
    second := readFasta(seqB)[0].seq
    n, m := len(first), len(second)

    // construct the alignment matrix (2D slice) and populate
    matches := makeMatrix(n, m)
    matchMatrix := populate(matches, first, second)
    sumList := sums(matchMatrix)

    // print out the matrix
    if *mat {
        printMatrix(matchMatrix)
    }

    // write scores
    writeScores(sumList)

    // parameters to stderr for debugging
    fmt.Fprintf(os.Stderr, "Running parameters:\nseqA: %v\n", *seqA)
    fmt.Fprintf(os.Stderr, "seqB: %v\n", *seqB)
    fmt.Fprintf(os.Stderr, "diag: %v\n", *diag)
    fmt.Fprintf(os.Stderr, "matrix: %v\n", *mat)
    fmt.Fprintf(os.Stderr, "penalty: %v\n", *penalty)
    fmt.Fprintf(os.Stderr, "out: %v\n", *out)

    // stop timer and print elapsed time to stderr
    stop := time.Now()
    elapsed := stop.Sub(start)
    fmt.Fprintf(os.Stderr, "Took %v to run.\n", elapsed)
}
