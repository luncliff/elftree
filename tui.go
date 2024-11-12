/*
 * ELF tree - Tree viewer for ELF library dependency
 *
 * Copyright (C) 2017  Namhyung Kim <namhyung@gmail.com>
 *
 * Released under MIT license.
 */
package main

import (
	"fmt"

	"github.com/gizak/termui"
)

type TreeItem struct {
	node   interface{}
	parent *TreeItem
	prev   *TreeItem // pointer to siblings
	next   *TreeItem
	child  *TreeItem // pointer to first child
	folded bool
	total  int // number of (shown) children (not count itself)
}

type TreeView struct {
	termui.Block // embedded
	Root         *TreeItem
	Top          *TreeItem
	Curr         *TreeItem

	ItemFgColor  termui.Attribute
	ItemBgColor  termui.Attribute
	FocusFgColor termui.Attribute
	FocusBgColor termui.Attribute

	idx int // current cursor position
	off int // first entry displayed
	pos int // position of x-axis

	rows int
	cols int
}

type FileInfo struct {
	Root *TreeItem
	Top  *TreeItem
	Curr *TreeItem
	idx  int
	off  int
	pos  int
}

const (
	MODE_FILE = iota
	MODE_SYMBOL
	MODE_DYNAMIC
	MODE_SECTION
)

var (
	mode  int
	finfo map[string]*FileInfo
	yinfo map[string]*FileInfo
	dinfo map[string]*FileInfo
	sinfo map[string]*FileInfo
	focus *TreeView
)

type StatusLine struct {
	termui.Block // embedded
	tv           *TreeView
}

func NewTreeView() *TreeView {
	tv := &TreeView{Block: *termui.NewBlock()}

	tv.ItemFgColor = termui.ThemeAttr("list.item.fg")
	tv.ItemBgColor = termui.ThemeAttr("list.item.bg")

	tv.idx = 0
	tv.off = 0
	return tv
}

func NewStatusLine(tv *TreeView) *StatusLine {
	sl := &StatusLine{Block: *termui.NewBlock()}

	sl.Block.Border = false

	sl.tv = tv
	return sl
}

func (ti *TreeItem) prevItem() *TreeItem {
	if ti.prev == nil {
		return ti.parent
	}

	ti = ti.prev

	// find last child of previous sibling
	for ti != nil {
		if ti.child == nil || ti.folded {
			return ti
		}

		ti = ti.child
		for ti.next != nil {
			ti = ti.next
		}
	}
	return nil
}

func (ti *TreeItem) nextItem() *TreeItem {
	if ti.child == nil || ti.folded {
		for ti != nil {
			if ti.next != nil {
				return ti.next
			}

			ti = ti.parent
		}
		return nil
	}
	return ti.child
}

func (ti *TreeItem) expand() {
	if !ti.folded || ti.child == nil {
		return
	}

	for c := ti.child; c != nil; c = c.next {
		ti.total += c.total + 1
	}

	for p := ti.parent; p != nil; p = p.parent {
		p.total += ti.total
	}

	ti.folded = false
}

func (ti *TreeItem) fold() {
	if ti.folded || ti.child == nil {
		return
	}

	for p := ti.parent; p != nil; p = p.parent {
		p.total -= ti.total
	}
	ti.total = 0

	ti.folded = true
}

func (ti *TreeItem) toggle() {
	if ti.folded {
		ti.expand()
	} else {
		ti.fold()
	}
}

func (tv *TreeView) drawDepsNode(buf termui.Buffer, dn *DepsNode, i, printed int, folded bool) {
	fg := tv.ItemFgColor
	bg := tv.ItemBgColor
	if i == tv.idx {
		if focus == tv {
			fg = tv.FocusFgColor
			bg = tv.FocusBgColor
		} else {
			fg = tv.ItemBgColor
			bg = tv.ItemFgColor
		}
	}

	indent := 3 * dn.depth
	text_width := tv.cols - 2 - indent

	if text_width < 0 {
		text_width = 0
	}

	cs := termui.DefaultTxBuilder.Build(dn.name, fg, bg)
	cs = termui.DTrimTxCls(cs, text_width)

	j := 0
	if i == tv.idx {
		// draw current line cursor from the beginning
		for j < indent {
			if j+1 > tv.pos {
				buf.Set(j+1-tv.pos, printed+1, termui.Cell{' ', fg, bg})
			}
			j++
		}
	} else {
		j = indent
	}

	if j+1 > tv.pos {
		if folded {
			buf.Set(j+1-tv.pos, printed+1, termui.Cell{'+', fg, bg})
		} else {
			buf.Set(j+1-tv.pos, printed+1, termui.Cell{'-', fg, bg})
		}
	}
	if j+2 > tv.pos {
		buf.Set(j+2-tv.pos, printed+1, termui.Cell{' ', fg, bg})
	}
	j += 2

	for _, vv := range cs {
		w := vv.Width()
		if j+1 > tv.pos {
			buf.Set(j+1-tv.pos, printed+1, vv)
		}
		j += w
	}

	if i != tv.idx {
		return
	}

	// draw current line cursor to the end
	for j < tv.cols+tv.pos {
		if j+1 > tv.pos {
			buf.Set(j+1-tv.pos, printed+1, termui.Cell{' ', fg, bg})
		}
		j++
	}
}

func (tv *TreeView) drawStrNode(buf termui.Buffer, s string, i, printed int, sign rune) {
	fg := tv.ItemFgColor
	bg := tv.ItemBgColor
	if i == tv.idx {
		if focus == tv {
			fg = tv.FocusFgColor
			bg = tv.FocusBgColor
		}
	}

	cs := termui.DefaultTxBuilder.Build(s, fg, bg)
	cs = termui.DTrimTxCls(cs, tv.cols-2)

	j := 0
	if j+1 > tv.pos {
		buf.Set(tv.X+j+1-tv.pos, printed+1, termui.Cell{sign, fg, bg})
	}
	if j+2 > tv.pos {
		buf.Set(tv.X+j+2-tv.pos, printed+1, termui.Cell{' ', fg, bg})
	}
	j += 2

	for _, vv := range cs {
		w := vv.Width()
		if j+1 > tv.pos {
			buf.Set(tv.X+j+1-tv.pos, printed+1, vv)
		}
		j += w
	}

	if i != tv.idx {
		return
	}

	// draw current line cursor to the end
	for j < tv.cols+tv.pos {
		if j+1 > tv.pos {
			buf.Set(tv.X+j+1-tv.pos, printed+1, termui.Cell{' ', fg, bg})
		}
		j++
	}
}

// Buffer implements Bufferer interface.
func (tv *TreeView) Buffer() termui.Buffer {
	buf := tv.Block.Buffer()

	i := 0
	printed := 0

	var ti *TreeItem
	for ti = tv.Root; ti != nil; ti = ti.nextItem() {
		if i < tv.off {
			i++
			continue
		}
		if printed == tv.rows {
			break
		}

		switch node := ti.node.(type) {
		case *DepsNode:
			tv.drawDepsNode(buf, node, i, printed, ti.folded)
			printed++
			i++
		case string:
			sign := ' '
			if ti.folded {
				sign = '+'
			} else if ti.child != nil {
				sign = '-'
			}
			tv.drawStrNode(buf, node, i, printed, sign)
			printed++
			i++
		default:
		}
	}

	return buf
}

func (tv *TreeView) Down() {
	if tv.idx < tv.Root.total {
		tv.idx++
		tv.Curr = tv.Curr.nextItem()
	}
	if tv.idx-tv.off >= tv.rows {
		tv.off++
		tv.Top = tv.Top.nextItem()
	}
}

func (tv *TreeView) Up() {
	if tv.idx > 0 {
		tv.idx--
		tv.Curr = tv.Curr.prevItem()
	}
	if tv.idx < tv.off {
		tv.off = tv.idx
		tv.Top = tv.Curr
	}
}

func (tv *TreeView) PageDown() {
	idx := tv.idx

	bottom := tv.off + tv.rows - 1
	if bottom > tv.Root.total {
		bottom = tv.Root.total
	}

	// At first, move to the bottom of current page
	if tv.idx != bottom {
		tv.idx = bottom

		for idx != bottom {
			tv.Curr = tv.Curr.nextItem()
			idx++
		}
		return
	}

	tv.idx += tv.rows
	if tv.idx > tv.Root.total {
		tv.idx = tv.Root.total
	}

	for idx != tv.idx {
		tv.Curr = tv.Curr.nextItem()
		idx++
	}

	off := tv.off

	if tv.idx-tv.off >= tv.rows {
		tv.off = tv.idx - tv.rows + 1

		for off != tv.off {
			tv.Top = tv.Top.nextItem()
			off++
		}
	}
}

func (tv *TreeView) PageUp() {
	idx := tv.idx

	// At first, move to the top of current page
	if tv.idx != tv.off {
		tv.idx = tv.off
		tv.Curr = tv.Top
		return
	}

	tv.idx -= tv.rows
	if tv.idx < 0 {
		tv.idx = 0
	}

	tv.off = tv.idx

	for idx != tv.idx {
		tv.Curr = tv.Curr.prevItem()
		idx--
	}

	tv.Top = tv.Curr
}

func (tv *TreeView) Home() {
	tv.idx = 0
	tv.off = 0
	tv.Curr = tv.Root
	tv.Top = tv.Root
}

func (tv *TreeView) End() {
	tv.idx = tv.Root.total
	tv.off = tv.idx - tv.rows + 1

	if tv.off < 0 {
		tv.off = 0
	}

	for next := tv.Curr; next != nil; next = next.nextItem() {
		tv.Curr = next
	}

	off := tv.idx
	top := tv.Curr

	for off != tv.off {
		top = top.prevItem()
		off--
	}

	tv.Top = top
}

func (tv *TreeView) Left(i int) {
	tv.pos -= i
	if tv.pos < 0 {
		tv.pos = 0
	}
}

func (tv *TreeView) Right(i int) {
	tv.pos += i
}

func (tv *TreeView) Toggle() {
	tv.Curr.toggle()
}

// Buffer implements Bufferer interface.
func (sl *StatusLine) Buffer() termui.Buffer {
	buf := sl.Block.Buffer()

	var line string

	curr := sl.tv.Curr
	if curr != nil {
		node := curr.node.(*DepsNode)
		line = node.name

		n := node.parent
		for n != nil {
			line = n.name + " > " + line

			n = n.parent
		}
	} else {
		line = "ELF tree"
	}

	fg := termui.ColorBlack
	bg := termui.ColorWhite

	cs := termui.DefaultTxBuilder.Build(line, fg, bg)
	cs = termui.DTrimTxCls(cs, sl.Width-3)

	buf.Set(0, sl.Y, termui.Cell{' ', fg, bg})
	buf.Set(1, sl.Y, termui.Cell{' ', fg, bg})

	j := 2
	for _, vv := range cs {
		w := vv.Width()
		buf.Set(j, sl.Y, vv)
		j += w
	}

	// draw status line to the end
	for j < sl.Width {
		buf.Set(j, sl.Y, termui.Cell{' ', fg, bg})
		j++
	}

	return buf
}

func makeDepsItems(dep *DepsNode, parent *TreeItem) *TreeItem {
	item := &TreeItem{node: dep, parent: parent, folded: false, total: len(dep.child)}

	var prev *TreeItem
	for _, v := range dep.child {
		c := makeDepsItems(v, item)

		if item.child == nil {
			item.child = c
		}
		if prev != nil {
			prev.next = c
			c.prev = prev
		}
		prev = c

		item.total += c.total
	}
	return item
}

func AddSubTree(name string, items []string, parent *TreeItem) {
	var t, p *TreeItem

	t = &TreeItem{node: name, parent: parent}

	if parent.child == nil {
		parent.child = t
	} else {
		p = parent.child
		for p.next != nil {
			p = p.next
		}

		p.next = t
		t.prev = p
	}

	p = nil
	parent = t
	for _, item := range items {
		t = &TreeItem{node: item, parent: parent}

		if p == nil {
			parent.child = t
		} else {
			p.next = t
			t.prev = p
		}

		p = t
	}

	parent.total = len(items)
	parent.parent.total += len(items) + 1
}

func makeFileInfo(name string, info *DepsInfo) *FileInfo {
	root := &TreeItem{node: name}

	// general file info
	AddSubTree("", nil, root)
	AddSubTree("File Info", []string{"  Path: " + info.path,
		"  Type: " + info.kind.String() + ", " + info.mach.String(),
		"  Data: " + info.bits.String() + ", " + info.endian.String()},
		root)

	// program headers
	var phdr []string
	for _, v := range info.prog {
		phdr = append(phdr, "  "+progHdrString(v))
	}
	AddSubTree("", nil, root)
	AddSubTree(fmt.Sprintf("%-16s   %s %18s  %10s  %8s",
		"Program Info", "Flags", "Vaddr", "Size", "Align"),
		phdr, root)

	// dependent libraries
	var libs []string
	for _, v := range info.libs {
		libs = append(libs, "  "+v)
	}
	AddSubTree("", nil, root)
	AddSubTree("Dependencies", libs, root)

	return &FileInfo{Root: root, Top: root, Curr: root}
}

func makeSymbolInfo(name string, info *DepsInfo) *FileInfo {
	root := &TreeItem{node: name}

	// dynamic symbols
	AddSubTree("", nil, root)
	var dsym []string
	for _, v := range info.dsym {
		dsym = append(dsym, makeSymbolString(v))
	}
	AddSubTree("Dynamic Symbols", dsym, root)

	// normal symbols
	AddSubTree("", nil, root)
	var nsym []string
	for _, v := range info.syms {
		nsym = append(nsym, makeSymbolString(v))
	}
	AddSubTree("Symbols", nsym, root)

	return &FileInfo{Root: root, Top: root, Curr: root}
}

func makeDynamicInfo(name string, info *DepsInfo) *FileInfo {
	root := &TreeItem{node: name}

	// dynamic info
	AddSubTree("", nil, root)
	AddSubTree("Dynamic Info", makeDynamicStrings(info), root)

	return &FileInfo{Root: root, Top: root, Curr: root}
}

func makeSectionInfo(name string, info *DepsInfo) *FileInfo {
	root := &TreeItem{node: name}

	// section headers
	AddSubTree("", nil, root)
	var sect []string
	sect = append(sect, fmt.Sprintf("  %4s %-24s %-12s %8s %8s %4s",
		"Idx", "Name", "Type", "Offset", "Size", "Flag"))
	for i, v := range info.sect {
		sect = append(sect, makeSectionString(i, v))
	}
	AddSubTree("Section Info", sect, root)

	return &FileInfo{Root: root, Top: root, Curr: root}
}

func saveInfoView(tv, iv *TreeView) {
	if focus != tv {
		return
	}

	curr := tv.Curr
	node := curr.node.(*DepsNode)

	var info *FileInfo

	if mode == MODE_FILE {
		info = finfo[node.name]
	} else if mode == MODE_SYMBOL {
		info = yinfo[node.name]
	} else if mode == MODE_DYNAMIC {
		info = dinfo[node.name]
	} else if mode == MODE_SECTION {
		info = sinfo[node.name]
	}

	info.Root = iv.Root
	info.Top = iv.Top
	info.Curr = iv.Curr

	info.off = iv.off
	info.idx = iv.idx
	info.pos = iv.pos
}

func restoreInfoView(tv, iv *TreeView) {
	if focus != tv {
		return
	}

	curr := tv.Curr
	node := curr.node.(*DepsNode)

	var info *FileInfo

	if mode == MODE_FILE {
		info = finfo[node.name]
	} else if mode == MODE_SYMBOL {
		info = yinfo[node.name]
	} else if mode == MODE_DYNAMIC {
		info = dinfo[node.name]
	} else if mode == MODE_SECTION {
		info = sinfo[node.name]
	}

	iv.Root = info.Root
	iv.Top = info.Top
	iv.Curr = info.Curr

	iv.off = info.off
	iv.idx = info.idx
	iv.pos = info.pos
}

func resize(tv, iv *TreeView, sl *StatusLine) {
	tv.Height = termui.TermHeight() - 1
	tv.Width = termui.TermWidth() * 3 / 5

	tv.rows = tv.Height - 2 // exclude border at top and bottom
	tv.cols = tv.Width - 2  // exclude border at left and right

	iv.Height = termui.TermHeight() - 1
	iv.Width = termui.TermWidth() - tv.Width
	iv.X = tv.Width

	iv.rows = iv.Height - 2
	iv.cols = iv.Width - 2

	sl.Height = 1
	sl.Width = termui.TermWidth()
	sl.Y = termui.TermHeight() - 1
}

func ShowWithTUI(dep *DepsNode) {
	if err := termui.Init(); err != nil {
		panic(err)
	}
	defer termui.Close()

	root := makeDepsItems(dep, nil)

	tv := NewTreeView()

	tv.Root = root
	tv.Curr = root
	tv.Top = root

	tv.FocusFgColor = termui.ColorYellow
	tv.FocusBgColor = termui.ColorBlue

	tv.BorderLabel = "ELF Tree"

	iv := NewTreeView()

	iv.FocusFgColor = termui.ColorYellow
	iv.FocusBgColor = termui.ColorBlue

	sl := NewStatusLine(tv)

	finfo = make(map[string]*FileInfo)
	yinfo = make(map[string]*FileInfo)
	dinfo = make(map[string]*FileInfo)
	sinfo = make(map[string]*FileInfo)

	for k, v := range deps {
		finfo[k] = makeFileInfo(k, &v)
		yinfo[k] = makeSymbolInfo(k, &v)
		dinfo[k] = makeDynamicInfo(k, &v)
		sinfo[k] = makeSectionInfo(k, &v)
	}
	mode = MODE_FILE
	focus = tv

	restoreInfoView(tv, iv)

	resize(tv, iv, sl)

	termui.Render(tv)
	termui.Render(iv)
	termui.Render(sl)

	// handle key pressing
	termui.Handle("/sys/kbd/q", func(termui.Event) {
		// press q to quit
		termui.StopLoop()
	})
	termui.Handle("/sys/kbd/C-c", func(termui.Event) {
		// press Ctrl-C to quit
		termui.StopLoop()
	})

	termui.Handle("/sys/kbd/f", func(termui.Event) {
		if focus == tv {
			mode = MODE_FILE
			restoreInfoView(tv, iv)
		}

		termui.Render(iv)
		termui.Render(sl)
	})
	termui.Handle("/sys/kbd/y", func(termui.Event) {
		if focus == tv {
			mode = MODE_SYMBOL
			restoreInfoView(tv, iv)
		}

		termui.Render(iv)
		termui.Render(sl)
	})
	termui.Handle("/sys/kbd/d", func(termui.Event) {
		if focus == tv {
			mode = MODE_DYNAMIC
			restoreInfoView(tv, iv)
		}

		termui.Render(iv)
		termui.Render(sl)
	})
	termui.Handle("/sys/kbd/s", func(termui.Event) {
		if focus == tv {
			mode = MODE_SECTION
			restoreInfoView(tv, iv)
		}

		termui.Render(iv)
		termui.Render(sl)
	})

	termui.Handle("/sys/kbd/<down>", func(termui.Event) {
		saveInfoView(tv, iv)
		focus.Down()
		restoreInfoView(tv, iv)

		termui.Render(tv)
		termui.Render(iv)
		termui.Render(sl)

	})
	termui.Handle("/sys/kbd/<up>", func(termui.Event) {
		saveInfoView(tv, iv)
		focus.Up()
		restoreInfoView(tv, iv)

		termui.Render(tv)
		termui.Render(iv)
		termui.Render(sl)

	})
	termui.Handle("/sys/kbd/<left>", func(termui.Event) {
		focus.Left(1)
		termui.Render(focus)
		// no need to redraw sl
	})
	termui.Handle("/sys/kbd/<right>", func(termui.Event) {
		focus.Right(1)
		termui.Render(focus)
		// no need to redraw sl
	})
	termui.Handle("/sys/kbd/<", func(termui.Event) {
		focus.Left(3)
		termui.Render(focus)
		// no need to redraw sl
	})
	termui.Handle("/sys/kbd/>", func(termui.Event) {
		focus.Right(3)
		termui.Render(focus)
		// no need to redraw sl
	})
	termui.Handle("/sys/kbd/<next>", func(termui.Event) {
		saveInfoView(tv, iv)
		focus.PageDown()
		restoreInfoView(tv, iv)

		termui.Render(tv)
		termui.Render(iv)
		termui.Render(sl)
	})
	termui.Handle("/sys/kbd/<previous>", func(termui.Event) {
		saveInfoView(tv, iv)
		focus.PageUp()
		restoreInfoView(tv, iv)

		termui.Render(tv)
		termui.Render(iv)
		termui.Render(sl)
	})
	termui.Handle("/sys/kbd/<home>", func(termui.Event) {
		saveInfoView(tv, iv)
		focus.Home()
		restoreInfoView(tv, iv)

		termui.Render(tv)
		termui.Render(iv)
		termui.Render(sl)
	})
	termui.Handle("/sys/kbd/<end>", func(termui.Event) {
		saveInfoView(tv, iv)
		focus.End()
		restoreInfoView(tv, iv)

		termui.Render(tv)
		termui.Render(iv)
		termui.Render(sl)
	})

	termui.Handle("/sys/kbd/<enter>", func(termui.Event) {
		focus.Toggle()
		termui.Render(tv)
		termui.Render(iv)
		termui.Render(sl)
	})

	termui.Handle("/sys/kbd/<tab>", func(termui.Event) {
		if focus == tv {
			focus = iv
		} else {
			focus = tv
		}

		termui.Render(tv)
		termui.Render(iv)
		termui.Render(sl)
	})

	termui.Handle("/sys/wnd/resize", func(termui.Event) {
		resize(tv, iv, sl)

		termui.Render(tv)
		termui.Render(iv)
		termui.Render(sl)
	})

	termui.Loop()
}
