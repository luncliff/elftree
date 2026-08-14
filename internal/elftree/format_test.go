/*
 * ELF tree - Tree viewer for ELF library dependency
 *
 * Copyright (C) 2017-2018  Namhyung Kim <namhyung@gmail.com>
 *
 * Released under MIT license.
 */
package elftree

import (
	"debug/elf"
	"strings"
	"testing"
)

// TestProgHdrString tests program header formatting
func TestProgHdrString(t *testing.T) {
	tests := []struct {
		name     string
		phdr     *elf.Prog
		expected string
	}{
		{
			name: "LOAD segment with RWX permissions",
			phdr: &elf.Prog{
				ProgHeader: elf.ProgHeader{
					Type:  elf.PT_LOAD,
					Flags: elf.PF_R | elf.PF_W | elf.PF_X,
					Vaddr: 0x400000,
					Memsz: 0x1000,
					Align: 0x1000,
				},
			},
			expected: "LOAD              RWX            0x400000      0x1000    0x1000",
		},
		{
			name: "LOAD segment with R-X permissions",
			phdr: &elf.Prog{
				ProgHeader: elf.ProgHeader{
					Type:  elf.PT_LOAD,
					Flags: elf.PF_R | elf.PF_X,
					Vaddr: 0x200000,
					Memsz: 0x500,
					Align: 0x200,
				},
			},
			expected: "LOAD              R_X            0x200000       0x500     0x200",
		},
		{
			name: "LOAD segment with RW- permissions",
			phdr: &elf.Prog{
				ProgHeader: elf.ProgHeader{
					Type:  elf.PT_LOAD,
					Flags: elf.PF_R | elf.PF_W,
					Vaddr: 0x600000,
					Memsz: 0x2000,
					Align: 0x1000,
				},
			},
			expected: "LOAD              RW_            0x600000      0x2000    0x1000",
		},
		{
			name: "GNU_STACK segment with R-X permissions",
			phdr: &elf.Prog{
				ProgHeader: elf.ProgHeader{
					Type:  GNU_STACK,
					Flags: elf.PF_R | elf.PF_X,
					Vaddr: 0,
					Memsz: 0,
					Align: 0x10,
				},
			},
			expected: "GNU_STACK         R_X                 0x0         0x0      0x10",
		},
		{
			name: "GNU_RELRO segment with R-- permissions",
			phdr: &elf.Prog{
				ProgHeader: elf.ProgHeader{
					Type:  GNU_RELRO,
					Flags: elf.PF_R,
					Vaddr: 0x500000,
					Memsz: 0x1000,
					Align: 0x1,
				},
			},
			expected: "GNU_RELRO         R__            0x500000      0x1000       0x1",
		},
		{
			name: "DYNAMIC segment",
			phdr: &elf.Prog{
				ProgHeader: elf.ProgHeader{
					Type:  elf.PT_DYNAMIC,
					Flags: elf.PF_R | elf.PF_W,
					Vaddr: 0x700000,
					Memsz: 0x200,
					Align: 0x8,
				},
			},
			expected: "DYNAMIC           RW_            0x700000       0x200       0x8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := progHdrString(tt.phdr)
			if result != tt.expected {
				t.Errorf("progHdrString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestStrFlags tests DT_FLAGS bitmask conversion
func TestStrFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    uint64
		expected string
	}{
		{
			name:     "no flags",
			flags:    0,
			expected: "",
		},
		{
			name:     "ORIGIN only",
			flags:    0x1,
			expected: "ORIGIN",
		},
		{
			name:     "SYMBOLIC only",
			flags:    0x2,
			expected: "SYMBOLIC",
		},
		{
			name:     "TEXTREL only",
			flags:    0x4,
			expected: "TEXTREL",
		},
		{
			name:     "BIND_NOW only",
			flags:    0x8,
			expected: "BIND_NOW",
		},
		{
			name:     "STATIC_TLS only",
			flags:    0x10,
			expected: "STATIC_TLS",
		},
		{
			name:     "ORIGIN | SYMBOLIC",
			flags:    0x1 | 0x2,
			expected: "ORIGIN|SYMBOLIC",
		},
		{
			name:     "TEXTREL | BIND_NOW",
			flags:    0x4 | 0x8,
			expected: "TEXTREL|BIND_NOW",
		},
		{
			name:     "all flags",
			flags:    0x1 | 0x2 | 0x4 | 0x8 | 0x10,
			expected: "ORIGIN|SYMBOLIC|TEXTREL|BIND_NOW|STATIC_TLS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strFlags(tt.flags)
			if result != tt.expected {
				t.Errorf("strFlags(%#x) = %q, want %q", tt.flags, result, tt.expected)
			}
		})
	}
}

// TestStrFlags1 tests DT_FLAGS_1 bitmask conversion
func TestStrFlags1(t *testing.T) {
	tests := []struct {
		name     string
		flags    uint64
		expected string
	}{
		{
			name:     "no flags",
			flags:    0,
			expected: "",
		},
		{
			name:     "NOW only",
			flags:    0x1,
			expected: "NOW",
		},
		{
			name:     "GLOBAL only",
			flags:    0x2,
			expected: "GLOBAL",
		},
		{
			name:     "NODELETE only",
			flags:    0x8,
			expected: "NODELETE",
		},
		{
			name:     "PIE only",
			flags:    0x2000000,
			expected: "SINGLETON",
		},
		{
			name:     "NOW | NODELETE",
			flags:    0x1 | 0x8,
			expected: "NOW|NODELETE",
		},
		{
			name:     "GLOBAL | NODELETE | INITFIRST",
			flags:    0x2 | 0x8 | 0x20,
			expected: "GLOBAL|NODELETE|INITFIRST",
		},
		{
			name:     "multiple common flags",
			flags:    0x1 | 0x2 | 0x4 | 0x8,
			expected: "NOW|GLOBAL|GROUP|NODELETE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strFlags1(tt.flags)
			if result != tt.expected {
				t.Errorf("strFlags1(%#x) = %q, want %q", tt.flags, result, tt.expected)
			}
		})
	}
}

// TestMakeSymbolString tests symbol formatting
func TestMakeSymbolString(t *testing.T) {
	tests := []struct {
		name     string
		sym      elf.Symbol
		expected string
	}{
		{
			name: "global function symbol",
			sym: elf.Symbol{
				Name:  "main",
				Info:  byte(elf.STB_GLOBAL)<<4 | byte(elf.STT_FUNC),
				Value: 0x1000,
			},
			expected: "      1000 FUN G main",
		},
		{
			name: "local object symbol",
			sym: elf.Symbol{
				Name:  "data",
				Info:  byte(elf.STB_LOCAL)<<4 | byte(elf.STT_OBJECT),
				Value: 0x2000,
			},
			expected: "      2000 OBJ L data",
		},
		{
			name: "weak function symbol",
			sym: elf.Symbol{
				Name:  "__gmon_start__",
				Info:  byte(elf.STB_WEAK)<<4 | byte(elf.STT_FUNC),
				Value: 0x0,
			},
			expected: "         0 FUN W __gmon_start__",
		},
		{
			name: "section symbol",
			sym: elf.Symbol{
				Name:  ".text",
				Info:  byte(elf.STB_LOCAL)<<4 | byte(elf.STT_SECTION),
				Value: 0x400,
			},
			expected: "       400 SEC L .text",
		},
		{
			name: "TLS symbol",
			sym: elf.Symbol{
				Name:  "errno",
				Info:  byte(elf.STB_GLOBAL)<<4 | byte(elf.STT_TLS),
				Value: 0x0,
			},
			expected: "         0 TLS G errno",
		},
		{
			name: "NOTYPE symbol",
			sym: elf.Symbol{
				Name:  "_start",
				Info:  byte(elf.STB_GLOBAL)<<4 | byte(elf.STT_NOTYPE),
				Value: 0x1050,
			},
			expected: "      1050 NON G _start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeSymbolString(tt.sym)
			if result != tt.expected {
				t.Errorf("makeSymbolString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestMakeSectionString tests section formatting
func TestMakeSectionString(t *testing.T) {
	tests := []struct {
		name     string
		idx      int
		sec      *elf.Section
		expected string
	}{
		{
			name: "writable allocated section",
			idx:  1,
			sec: &elf.Section{
				SectionHeader: elf.SectionHeader{
					Name:   ".data",
					Type:   elf.SHT_PROGBITS,
					Flags:  elf.SHF_WRITE | elf.SHF_ALLOC,
					Offset: 0x2000,
					Size:   0x100,
				},
			},
			expected: "  [ 1] .data                    PROGBITS         2000      100   WA",
		},
		{
			name: "executable allocated section",
			idx:  2,
			sec: &elf.Section{
				SectionHeader: elf.SectionHeader{
					Name:   ".text",
					Type:   elf.SHT_PROGBITS,
					Flags:  elf.SHF_ALLOC | elf.SHF_EXECINSTR,
					Offset: 0x1000,
					Size:   0x500,
				},
			},
			expected: "  [ 2] .text                    PROGBITS         1000      500   AX",
		},
		{
			name: "read-only allocated section",
			idx:  3,
			sec: &elf.Section{
				SectionHeader: elf.SectionHeader{
					Name:   ".rodata",
					Type:   elf.SHT_PROGBITS,
					Flags:  elf.SHF_ALLOC,
					Offset: 0x1500,
					Size:   0x200,
				},
			},
			expected: "  [ 3] .rodata                  PROGBITS         1500      200    A",
		},
		{
			name: "string table section",
			idx:  4,
			sec: &elf.Section{
				SectionHeader: elf.SectionHeader{
					Name:   ".strtab",
					Type:   elf.SHT_STRTAB,
					Flags:  0,
					Offset: 0x3000,
					Size:   0x50,
				},
			},
			expected: "  [ 4] .strtab                  STRTAB           3000       50     ",
		},
		{
			name: "symbol table section",
			idx:  5,
			sec: &elf.Section{
				SectionHeader: elf.SectionHeader{
					Name:   ".symtab",
					Type:   elf.SHT_SYMTAB,
					Flags:  0,
					Offset: 0x4000,
					Size:   0x300,
				},
			},
			expected: "  [ 5] .symtab                  SYMTAB           4000      300     ",
		},
		{
			name: "dynamic section with all flags",
			idx:  6,
			sec: &elf.Section{
				SectionHeader: elf.SectionHeader{
					Name:   ".dynamic",
					Type:   elf.SHT_DYNAMIC,
					Flags:  elf.SHF_WRITE | elf.SHF_ALLOC,
					Offset: 0x5000,
					Size:   0x150,
				},
			},
			expected: "  [ 6] .dynamic                 DYNAMIC          5000      150   WA",
		},
		{
			name: "bss section",
			idx:  7,
			sec: &elf.Section{
				SectionHeader: elf.SectionHeader{
					Name:   ".bss",
					Type:   elf.SHT_NOBITS,
					Flags:  elf.SHF_WRITE | elf.SHF_ALLOC,
					Offset: 0x6000,
					Size:   0x800,
				},
			},
			expected: "  [ 7] .bss                     NOBITS           6000      800   WA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeSectionString(tt.idx, tt.sec)
			if result != tt.expected {
				t.Errorf("makeSectionString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestMakeDynamicStrings tests dynamic section entry formatting
func TestMakeDynamicStrings(t *testing.T) {
	tests := []struct {
		name     string
		info     *DepsInfo
		expected []string
	}{
		{
			name: "DT_NEEDED entries",
			info: &DepsInfo{
				dyns: []DynInfo{
					{tag: elf.DT_NEEDED, val: "libc.so.6"},
					{tag: elf.DT_NEEDED, val: "libm.so.6"},
				},
			},
			expected: []string{
				"  DT_NEEDED         libc.so.6",
				"  DT_NEEDED         libm.so.6",
			},
		},
		{
			name: "DT_RPATH and DT_RUNPATH",
			info: &DepsInfo{
				dyns: []DynInfo{
					{tag: elf.DT_RPATH, val: "/usr/local/lib"},
					{tag: elf.DT_RUNPATH, val: "/opt/lib"},
				},
			},
			expected: []string{
				"  DT_RPATH          /usr/local/lib",
				"  DT_RUNPATH        /opt/lib",
			},
		},
		{
			name: "DT_SONAME",
			info: &DepsInfo{
				dyns: []DynInfo{
					{tag: elf.DT_SONAME, val: "libtest.so.1"},
				},
			},
			expected: []string{
				"  DT_SONAME         libtest.so.1",
			},
		},
		{
			name: "DT_FLAGS with multiple bits",
			info: &DepsInfo{
				dyns: []DynInfo{
					{tag: elf.DT_FLAGS, val: uint64(0x1 | 0x8)}, // ORIGIN | BIND_NOW
				},
			},
			expected: []string{
				"  DT_FLAGS          ORIGIN|BIND_NOW",
			},
		},
		{
			name: "DT_FLAGS_1 with NOW and NODELETE",
			info: &DepsInfo{
				dyns: []DynInfo{
					{tag: DT_FLAGS_1, val: uint64(0x1 | 0x8)}, // NOW | NODELETE
				},
			},
			expected: []string{
				"  DT_FLAGS_1        NOW|NODELETE",
			},
		},
		{
			name: "GNU_HASH and counters",
			info: &DepsInfo{
				dyns: []DynInfo{
					{tag: DT_GNU_HASH, val: uint64(0x400)},
					{tag: DT_RELACOUNT, val: uint64(5)},
					{tag: DT_RELCOUNT, val: uint64(10)},
				},
			},
			expected: []string{
				"  DT_GNU_HASH       400",
				"  DT_RELACOUNT      5",
				"  DT_RELCOUNT       10",
			},
		},
		{
			name: "version information",
			info: &DepsInfo{
				dyns: []DynInfo{
					{tag: DT_VERDEF, val: uint64(0x1000)},
					{tag: DT_VERDEFNUM, val: uint64(2)},
					{tag: DT_VERNEED, val: uint64(0x2000)},
					{tag: DT_VERNEEDNUM, val: uint64(3)},
				},
			},
			expected: []string{
				"  DT_VERDEF         1000",
				"  DT_VERDEFNUM      2",
				"  DT_VERNEED        2000",
				"  DT_VERNEEDNUM     3",
			},
		},
		{
			name: "mixed dynamic entries",
			info: &DepsInfo{
				dyns: []DynInfo{
					{tag: elf.DT_NEEDED, val: "libc.so.6"},
					{tag: elf.DT_FLAGS, val: uint64(0x8)}, // BIND_NOW
					{tag: DT_GNU_HASH, val: uint64(0x300)},
					{tag: elf.DT_SONAME, val: "libfoo.so.1"},
				},
			},
			expected: []string{
				"  DT_NEEDED         libc.so.6",
				"  DT_FLAGS          BIND_NOW",
				"  DT_GNU_HASH       300",
				"  DT_SONAME         libfoo.so.1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeDynamicStrings(tt.info)
			if len(result) != len(tt.expected) {
				t.Errorf("makeDynamicStrings() returned %d entries, want %d", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("entry %d: got %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// Benchmark tests for performance tracking
func BenchmarkStrFlags(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strFlags(0x1 | 0x2 | 0x4 | 0x8 | 0x10)
	}
}

func BenchmarkStrFlags1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strFlags1(0x1 | 0x2 | 0x4 | 0x8)
	}
}

func BenchmarkProgHdrString(b *testing.B) {
	phdr := &elf.Prog{
		ProgHeader: elf.ProgHeader{
			Type:  elf.PT_LOAD,
			Flags: elf.PF_R | elf.PF_W | elf.PF_X,
			Vaddr: 0x400000,
			Memsz: 0x1000,
			Align: 0x1000,
		},
	}
	for i := 0; i < b.N; i++ {
		progHdrString(phdr)
	}
}

func BenchmarkMakeSymbolString(b *testing.B) {
	sym := elf.Symbol{
		Name:  "main",
		Info:  byte(elf.STB_GLOBAL)<<4 | byte(elf.STT_FUNC),
		Value: 0x1000,
	}
	for i := 0; i < b.N; i++ {
		makeSymbolString(sym)
	}
}

func BenchmarkMakeSectionString(b *testing.B) {
	sec := &elf.Section{
		SectionHeader: elf.SectionHeader{
			Name:   ".text",
			Type:   elf.SHT_PROGBITS,
			Flags:  elf.SHF_ALLOC | elf.SHF_EXECINSTR,
			Offset: 0x1000,
			Size:   0x500,
		},
	}
	for i := 0; i < b.N; i++ {
		makeSectionString(1, sec)
	}
}

func BenchmarkMakeDynamicStrings(b *testing.B) {
	info := &DepsInfo{
		dyns: []DynInfo{
			{tag: elf.DT_NEEDED, val: "libc.so.6"},
			{tag: elf.DT_FLAGS, val: uint64(0x8)},
			{tag: DT_GNU_HASH, val: uint64(0x300)},
		},
	}
	for i := 0; i < b.N; i++ {
		_ = makeDynamicStrings(info)
	}
}

// TestFlagConstants verifies that flag constants are defined correctly
func TestFlagConstants(t *testing.T) {
	// Verify custom DT_FLAGS_1 constants
	if DT_FLAGS_1 != elf.DT_VERSYM+11 {
		t.Errorf("DT_FLAGS_1 = %#x, want %#x", DT_FLAGS_1, elf.DT_VERSYM+11)
	}

	// Verify GNU program header constants
	expectedGnuEhFrame := elf.PT_LOOS + 74769744
	if GNU_EH_FRAME != expectedGnuEhFrame {
		t.Errorf("GNU_EH_FRAME = %#x, want %#x", GNU_EH_FRAME, expectedGnuEhFrame)
	}
}

// TestSectionFlagCombinations tests various section flag combinations
func TestSectionFlagCombinations(t *testing.T) {
	tests := []struct {
		name  string
		flags elf.SectionFlag
		want  string
	}{

		{"write + alloc + exec", elf.SHF_WRITE | elf.SHF_ALLOC | elf.SHF_EXECINSTR, "WAX"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sec := &elf.Section{
				SectionHeader: elf.SectionHeader{
					Name:   ".test",
					Type:   elf.SHT_PROGBITS,
					Flags:  tt.flags,
					Offset: 0,
					Size:   0,
				},
			}
			result := makeSectionString(0, sec)
			// Extract just the flags part (last field)
			parts := strings.Fields(result)
			if len(parts) < 6 {
				t.Fatalf("unexpected format: %q", result)
			}
			got := strings.TrimSpace(parts[len(parts)-1])
			if got != tt.want {
				t.Errorf("flags = %q, want %q", got, tt.want)
			}
		})
	}
}
