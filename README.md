# goAlign
R2C2 Smith-Waterman aligner that outputs a list of summed scores

### Dependencies ###
- [Go](https://golang.org/dl/)

To compile, use make
```bash
make
```

### Usage ###
goAlign takes two sequences, populates a score matrix for the alignment, and sums scores parallel to the main diagonal.

Options:

  -a    First fasta file to align (required)

  -b    Second fasta file to align (required)

  -d    Use to exclude the main diagonal (it will be set to zero)

  -m    Use to output the entire score matrix to stdout

  -p    Penalty for either gap open or extend (defaults to 25)

  -o    Output file name (defaults to "SW_PARSE.txt")

To align and do nothing else:
```bash
./goAlign -a a.fasta -b b.fasta -p 25
```

To align and exclude the main diagonal:
```bash
./goAlign -a a.fasta -b b.fasta -d
```

To align and output the entire score matrix to matrix.mat:
```bash
./goAlign -a a.fasta -b b.fasta -m >matrix.mat
```

Run parameters and a timer are piped to stderr
