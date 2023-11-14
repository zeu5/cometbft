// Package slot is generated by gogll. Do not edit.
package slot

import (
	"bytes"
	"fmt"

	"github.com/zeu5/cometbft/test/e2e/pkg/grammar/recovery/grammar-auto/parser/symbols"
)

type Label int

const (
	Commit0R0 Label = iota
	Commit0R1
	ConsensusExec0R0
	ConsensusExec0R1
	ConsensusHeight0R0
	ConsensusHeight0R1
	ConsensusHeight0R2
	ConsensusHeight0R3
	ConsensusHeight1R0
	ConsensusHeight1R1
	ConsensusHeight1R2
	ConsensusHeights0R0
	ConsensusHeights0R1
	ConsensusHeights1R0
	ConsensusHeights1R1
	ConsensusHeights1R2
	ConsensusRound0R0
	ConsensusRound0R1
	ConsensusRound1R0
	ConsensusRound1R1
	ConsensusRounds0R0
	ConsensusRounds0R1
	ConsensusRounds1R0
	ConsensusRounds1R1
	ConsensusRounds1R2
	FinalizeBlock0R0
	FinalizeBlock0R1
	NonProposer0R0
	NonProposer0R1
	PrepareProposal0R0
	PrepareProposal0R1
	ProcessProposal0R0
	ProcessProposal0R1
	Proposer0R0
	Proposer0R1
	Proposer1R0
	Proposer1R1
	Proposer1R2
	Recovery0R0
	Recovery0R1
	Start0R0
	Start0R1
)

type Slot struct {
	NT      symbols.NT
	Alt     int
	Pos     int
	Symbols symbols.Symbols
	Label   Label
}

type Index struct {
	NT  symbols.NT
	Alt int
	Pos int
}

func GetAlternates(nt symbols.NT) []Label {
	alts, exist := alternates[nt]
	if !exist {
		panic(fmt.Sprintf("Invalid NT %s", nt))
	}
	return alts
}

func GetLabel(nt symbols.NT, alt, pos int) Label {
	l, exist := slotIndex[Index{nt, alt, pos}]
	if exist {
		return l
	}
	panic(fmt.Sprintf("Error: no slot label for NT=%s, alt=%d, pos=%d", nt, alt, pos))
}

func (l Label) EoR() bool {
	return l.Slot().EoR()
}

func (l Label) Head() symbols.NT {
	return l.Slot().NT
}

func (l Label) Index() Index {
	s := l.Slot()
	return Index{s.NT, s.Alt, s.Pos}
}

func (l Label) Alternate() int {
	return l.Slot().Alt
}

func (l Label) Pos() int {
	return l.Slot().Pos
}

func (l Label) Slot() *Slot {
	s, exist := slots[l]
	if !exist {
		panic(fmt.Sprintf("Invalid slot label %d", l))
	}
	return s
}

func (l Label) String() string {
	return l.Slot().String()
}

func (l Label) Symbols() symbols.Symbols {
	return l.Slot().Symbols
}

func (s *Slot) EoR() bool {
	return s.Pos >= len(s.Symbols)
}

func (s *Slot) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%s : ", s.NT)
	for i, sym := range s.Symbols {
		if i == s.Pos {
			fmt.Fprintf(buf, "∙")
		}
		fmt.Fprintf(buf, "%s ", sym)
	}
	if s.Pos >= len(s.Symbols) {
		fmt.Fprintf(buf, "∙")
	}
	return buf.String()
}

var slots = map[Label]*Slot{
	Commit0R0: {
		symbols.NT_Commit, 0, 0,
		symbols.Symbols{
			symbols.T_0,
		},
		Commit0R0,
	},
	Commit0R1: {
		symbols.NT_Commit, 0, 1,
		symbols.Symbols{
			symbols.T_0,
		},
		Commit0R1,
	},
	ConsensusExec0R0: {
		symbols.NT_ConsensusExec, 0, 0,
		symbols.Symbols{
			symbols.NT_ConsensusHeights,
		},
		ConsensusExec0R0,
	},
	ConsensusExec0R1: {
		symbols.NT_ConsensusExec, 0, 1,
		symbols.Symbols{
			symbols.NT_ConsensusHeights,
		},
		ConsensusExec0R1,
	},
	ConsensusHeight0R0: {
		symbols.NT_ConsensusHeight, 0, 0,
		symbols.Symbols{
			symbols.NT_ConsensusRounds,
			symbols.NT_FinalizeBlock,
			symbols.NT_Commit,
		},
		ConsensusHeight0R0,
	},
	ConsensusHeight0R1: {
		symbols.NT_ConsensusHeight, 0, 1,
		symbols.Symbols{
			symbols.NT_ConsensusRounds,
			symbols.NT_FinalizeBlock,
			symbols.NT_Commit,
		},
		ConsensusHeight0R1,
	},
	ConsensusHeight0R2: {
		symbols.NT_ConsensusHeight, 0, 2,
		symbols.Symbols{
			symbols.NT_ConsensusRounds,
			symbols.NT_FinalizeBlock,
			symbols.NT_Commit,
		},
		ConsensusHeight0R2,
	},
	ConsensusHeight0R3: {
		symbols.NT_ConsensusHeight, 0, 3,
		symbols.Symbols{
			symbols.NT_ConsensusRounds,
			symbols.NT_FinalizeBlock,
			symbols.NT_Commit,
		},
		ConsensusHeight0R3,
	},
	ConsensusHeight1R0: {
		symbols.NT_ConsensusHeight, 1, 0,
		symbols.Symbols{
			symbols.NT_FinalizeBlock,
			symbols.NT_Commit,
		},
		ConsensusHeight1R0,
	},
	ConsensusHeight1R1: {
		symbols.NT_ConsensusHeight, 1, 1,
		symbols.Symbols{
			symbols.NT_FinalizeBlock,
			symbols.NT_Commit,
		},
		ConsensusHeight1R1,
	},
	ConsensusHeight1R2: {
		symbols.NT_ConsensusHeight, 1, 2,
		symbols.Symbols{
			symbols.NT_FinalizeBlock,
			symbols.NT_Commit,
		},
		ConsensusHeight1R2,
	},
	ConsensusHeights0R0: {
		symbols.NT_ConsensusHeights, 0, 0,
		symbols.Symbols{
			symbols.NT_ConsensusHeight,
		},
		ConsensusHeights0R0,
	},
	ConsensusHeights0R1: {
		symbols.NT_ConsensusHeights, 0, 1,
		symbols.Symbols{
			symbols.NT_ConsensusHeight,
		},
		ConsensusHeights0R1,
	},
	ConsensusHeights1R0: {
		symbols.NT_ConsensusHeights, 1, 0,
		symbols.Symbols{
			symbols.NT_ConsensusHeight,
			symbols.NT_ConsensusHeights,
		},
		ConsensusHeights1R0,
	},
	ConsensusHeights1R1: {
		symbols.NT_ConsensusHeights, 1, 1,
		symbols.Symbols{
			symbols.NT_ConsensusHeight,
			symbols.NT_ConsensusHeights,
		},
		ConsensusHeights1R1,
	},
	ConsensusHeights1R2: {
		symbols.NT_ConsensusHeights, 1, 2,
		symbols.Symbols{
			symbols.NT_ConsensusHeight,
			symbols.NT_ConsensusHeights,
		},
		ConsensusHeights1R2,
	},
	ConsensusRound0R0: {
		symbols.NT_ConsensusRound, 0, 0,
		symbols.Symbols{
			symbols.NT_Proposer,
		},
		ConsensusRound0R0,
	},
	ConsensusRound0R1: {
		symbols.NT_ConsensusRound, 0, 1,
		symbols.Symbols{
			symbols.NT_Proposer,
		},
		ConsensusRound0R1,
	},
	ConsensusRound1R0: {
		symbols.NT_ConsensusRound, 1, 0,
		symbols.Symbols{
			symbols.NT_NonProposer,
		},
		ConsensusRound1R0,
	},
	ConsensusRound1R1: {
		symbols.NT_ConsensusRound, 1, 1,
		symbols.Symbols{
			symbols.NT_NonProposer,
		},
		ConsensusRound1R1,
	},
	ConsensusRounds0R0: {
		symbols.NT_ConsensusRounds, 0, 0,
		symbols.Symbols{
			symbols.NT_ConsensusRound,
		},
		ConsensusRounds0R0,
	},
	ConsensusRounds0R1: {
		symbols.NT_ConsensusRounds, 0, 1,
		symbols.Symbols{
			symbols.NT_ConsensusRound,
		},
		ConsensusRounds0R1,
	},
	ConsensusRounds1R0: {
		symbols.NT_ConsensusRounds, 1, 0,
		symbols.Symbols{
			symbols.NT_ConsensusRound,
			symbols.NT_ConsensusRounds,
		},
		ConsensusRounds1R0,
	},
	ConsensusRounds1R1: {
		symbols.NT_ConsensusRounds, 1, 1,
		symbols.Symbols{
			symbols.NT_ConsensusRound,
			symbols.NT_ConsensusRounds,
		},
		ConsensusRounds1R1,
	},
	ConsensusRounds1R2: {
		symbols.NT_ConsensusRounds, 1, 2,
		symbols.Symbols{
			symbols.NT_ConsensusRound,
			symbols.NT_ConsensusRounds,
		},
		ConsensusRounds1R2,
	},
	FinalizeBlock0R0: {
		symbols.NT_FinalizeBlock, 0, 0,
		symbols.Symbols{
			symbols.T_1,
		},
		FinalizeBlock0R0,
	},
	FinalizeBlock0R1: {
		symbols.NT_FinalizeBlock, 0, 1,
		symbols.Symbols{
			symbols.T_1,
		},
		FinalizeBlock0R1,
	},
	NonProposer0R0: {
		symbols.NT_NonProposer, 0, 0,
		symbols.Symbols{
			symbols.NT_ProcessProposal,
		},
		NonProposer0R0,
	},
	NonProposer0R1: {
		symbols.NT_NonProposer, 0, 1,
		symbols.Symbols{
			symbols.NT_ProcessProposal,
		},
		NonProposer0R1,
	},
	PrepareProposal0R0: {
		symbols.NT_PrepareProposal, 0, 0,
		symbols.Symbols{
			symbols.T_2,
		},
		PrepareProposal0R0,
	},
	PrepareProposal0R1: {
		symbols.NT_PrepareProposal, 0, 1,
		symbols.Symbols{
			symbols.T_2,
		},
		PrepareProposal0R1,
	},
	ProcessProposal0R0: {
		symbols.NT_ProcessProposal, 0, 0,
		symbols.Symbols{
			symbols.T_3,
		},
		ProcessProposal0R0,
	},
	ProcessProposal0R1: {
		symbols.NT_ProcessProposal, 0, 1,
		symbols.Symbols{
			symbols.T_3,
		},
		ProcessProposal0R1,
	},
	Proposer0R0: {
		symbols.NT_Proposer, 0, 0,
		symbols.Symbols{
			symbols.NT_PrepareProposal,
		},
		Proposer0R0,
	},
	Proposer0R1: {
		symbols.NT_Proposer, 0, 1,
		symbols.Symbols{
			symbols.NT_PrepareProposal,
		},
		Proposer0R1,
	},
	Proposer1R0: {
		symbols.NT_Proposer, 1, 0,
		symbols.Symbols{
			symbols.NT_PrepareProposal,
			symbols.NT_ProcessProposal,
		},
		Proposer1R0,
	},
	Proposer1R1: {
		symbols.NT_Proposer, 1, 1,
		symbols.Symbols{
			symbols.NT_PrepareProposal,
			symbols.NT_ProcessProposal,
		},
		Proposer1R1,
	},
	Proposer1R2: {
		symbols.NT_Proposer, 1, 2,
		symbols.Symbols{
			symbols.NT_PrepareProposal,
			symbols.NT_ProcessProposal,
		},
		Proposer1R2,
	},
	Recovery0R0: {
		symbols.NT_Recovery, 0, 0,
		symbols.Symbols{
			symbols.NT_ConsensusExec,
		},
		Recovery0R0,
	},
	Recovery0R1: {
		symbols.NT_Recovery, 0, 1,
		symbols.Symbols{
			symbols.NT_ConsensusExec,
		},
		Recovery0R1,
	},
	Start0R0: {
		symbols.NT_Start, 0, 0,
		symbols.Symbols{
			symbols.NT_Recovery,
		},
		Start0R0,
	},
	Start0R1: {
		symbols.NT_Start, 0, 1,
		symbols.Symbols{
			symbols.NT_Recovery,
		},
		Start0R1,
	},
}

var slotIndex = map[Index]Label{
	Index{symbols.NT_Commit, 0, 0}:           Commit0R0,
	Index{symbols.NT_Commit, 0, 1}:           Commit0R1,
	Index{symbols.NT_ConsensusExec, 0, 0}:    ConsensusExec0R0,
	Index{symbols.NT_ConsensusExec, 0, 1}:    ConsensusExec0R1,
	Index{symbols.NT_ConsensusHeight, 0, 0}:  ConsensusHeight0R0,
	Index{symbols.NT_ConsensusHeight, 0, 1}:  ConsensusHeight0R1,
	Index{symbols.NT_ConsensusHeight, 0, 2}:  ConsensusHeight0R2,
	Index{symbols.NT_ConsensusHeight, 0, 3}:  ConsensusHeight0R3,
	Index{symbols.NT_ConsensusHeight, 1, 0}:  ConsensusHeight1R0,
	Index{symbols.NT_ConsensusHeight, 1, 1}:  ConsensusHeight1R1,
	Index{symbols.NT_ConsensusHeight, 1, 2}:  ConsensusHeight1R2,
	Index{symbols.NT_ConsensusHeights, 0, 0}: ConsensusHeights0R0,
	Index{symbols.NT_ConsensusHeights, 0, 1}: ConsensusHeights0R1,
	Index{symbols.NT_ConsensusHeights, 1, 0}: ConsensusHeights1R0,
	Index{symbols.NT_ConsensusHeights, 1, 1}: ConsensusHeights1R1,
	Index{symbols.NT_ConsensusHeights, 1, 2}: ConsensusHeights1R2,
	Index{symbols.NT_ConsensusRound, 0, 0}:   ConsensusRound0R0,
	Index{symbols.NT_ConsensusRound, 0, 1}:   ConsensusRound0R1,
	Index{symbols.NT_ConsensusRound, 1, 0}:   ConsensusRound1R0,
	Index{symbols.NT_ConsensusRound, 1, 1}:   ConsensusRound1R1,
	Index{symbols.NT_ConsensusRounds, 0, 0}:  ConsensusRounds0R0,
	Index{symbols.NT_ConsensusRounds, 0, 1}:  ConsensusRounds0R1,
	Index{symbols.NT_ConsensusRounds, 1, 0}:  ConsensusRounds1R0,
	Index{symbols.NT_ConsensusRounds, 1, 1}:  ConsensusRounds1R1,
	Index{symbols.NT_ConsensusRounds, 1, 2}:  ConsensusRounds1R2,
	Index{symbols.NT_FinalizeBlock, 0, 0}:    FinalizeBlock0R0,
	Index{symbols.NT_FinalizeBlock, 0, 1}:    FinalizeBlock0R1,
	Index{symbols.NT_NonProposer, 0, 0}:      NonProposer0R0,
	Index{symbols.NT_NonProposer, 0, 1}:      NonProposer0R1,
	Index{symbols.NT_PrepareProposal, 0, 0}:  PrepareProposal0R0,
	Index{symbols.NT_PrepareProposal, 0, 1}:  PrepareProposal0R1,
	Index{symbols.NT_ProcessProposal, 0, 0}:  ProcessProposal0R0,
	Index{symbols.NT_ProcessProposal, 0, 1}:  ProcessProposal0R1,
	Index{symbols.NT_Proposer, 0, 0}:         Proposer0R0,
	Index{symbols.NT_Proposer, 0, 1}:         Proposer0R1,
	Index{symbols.NT_Proposer, 1, 0}:         Proposer1R0,
	Index{symbols.NT_Proposer, 1, 1}:         Proposer1R1,
	Index{symbols.NT_Proposer, 1, 2}:         Proposer1R2,
	Index{symbols.NT_Recovery, 0, 0}:         Recovery0R0,
	Index{symbols.NT_Recovery, 0, 1}:         Recovery0R1,
	Index{symbols.NT_Start, 0, 0}:            Start0R0,
	Index{symbols.NT_Start, 0, 1}:            Start0R1,
}

var alternates = map[symbols.NT][]Label{
	symbols.NT_Start:            []Label{Start0R0},
	symbols.NT_Recovery:         []Label{Recovery0R0},
	symbols.NT_ConsensusExec:    []Label{ConsensusExec0R0},
	symbols.NT_ConsensusHeights: []Label{ConsensusHeights0R0, ConsensusHeights1R0},
	symbols.NT_ConsensusHeight:  []Label{ConsensusHeight0R0, ConsensusHeight1R0},
	symbols.NT_ConsensusRounds:  []Label{ConsensusRounds0R0, ConsensusRounds1R0},
	symbols.NT_ConsensusRound:   []Label{ConsensusRound0R0, ConsensusRound1R0},
	symbols.NT_Proposer:         []Label{Proposer0R0, Proposer1R0},
	symbols.NT_NonProposer:      []Label{NonProposer0R0},
	symbols.NT_FinalizeBlock:    []Label{FinalizeBlock0R0},
	symbols.NT_Commit:           []Label{Commit0R0},
	symbols.NT_PrepareProposal:  []Label{PrepareProposal0R0},
	symbols.NT_ProcessProposal:  []Label{ProcessProposal0R0},
}
