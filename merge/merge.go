// Package merge provides for the merging of two PDFContexts.
//
// This means the concatenation of two page trees and merging all data involved.
package merge

import (
	"io/ioutil"
	"log"
	"os"
	"sort"

	"github.com/hhrutter/pdfcpu/types"
)

var logDebugMerge, logInfoMerge, logErrorMerge *log.Logger

func init() {
	//logDebugMerge = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	logDebugMerge = log.New(ioutil.Discard, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	logInfoMerge = log.New(ioutil.Discard, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	logErrorMerge = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Verbose controls logging output.
func Verbose(verbose bool) {
	out := ioutil.Discard
	if verbose {
		out = os.Stdout
	}
	logInfoMerge = log.New(out, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func patchIndRef(indRef *types.PDFIndirectRef, lookup map[int]int) {
	i := indRef.ObjectNumber.Value()
	indRef.ObjectNumber = types.PDFInteger(lookup[i])
}

func patchObject(o interface{}, lookup map[int]int) interface{} {

	logDebugMerge.Printf("patchObject before: %v\n", o)

	var ob interface{}

	switch obj := o.(type) {

	case types.PDFIndirectRef:
		patchIndRef(&obj, lookup)
		ob = obj

	case types.PDFDict:
		patchDict(&obj, lookup)
		ob = obj

	case types.PDFStreamDict:
		patchDict(&obj.PDFDict, lookup)
		ob = obj

	case types.PDFObjectStreamDict:
		patchDict(&obj.PDFDict, lookup)
		ob = obj

	case types.PDFXRefStreamDict:
		patchDict(&obj.PDFDict, lookup)
		ob = obj

	case types.PDFArray:
		patchArray(&obj, lookup)
		ob = obj

	}

	logDebugMerge.Printf("patchObject end: %v\n", ob)

	return ob
}

func patchDict(dict *types.PDFDict, lookup map[int]int) {

	logDebugMerge.Printf("patchDict before: %v\n", dict)

	for k, obj := range dict.Dict {
		o := patchObject(obj, lookup)
		if o != nil {
			dict.Dict[k] = o
		}
	}

	logDebugMerge.Printf("patchDict after: %v\n", dict)
}

func patchArray(arr *types.PDFArray, lookup map[int]int) {

	logDebugMerge.Printf("patchArray begin: %v\n", arr)

	for i, obj := range *arr {
		o := patchObject(obj, lookup)
		if o != nil {
			(*arr)[i] = o
		}
	}

	logDebugMerge.Printf("patchArray end: %v\n", arr)
}

func sortedKeys(ctx *types.PDFContext) []int {

	var keys []int

	for k := range ctx.Table {
		if k == 0 {
			// obj#0 is always the head of the freelist.
			continue
		}
		keys = append(keys, k)
	}

	sort.Ints(keys)

	return keys
}

func lookupTable(keys []int, i int) map[int]int {

	m := map[int]int{}

	for _, k := range keys {
		m[k] = i
		i++
	}

	return m
}

func patchObjects(s types.IntSet, lookup map[int]int) types.IntSet {

	t := types.IntSet{}

	for k, v := range s {
		if v {
			t[lookup[k]] = v
		}
	}

	return t
}

func patchSourceObjectNumbers(ctxSource, ctxDest *types.PDFContext) {

	logInfoMerge.Printf("patchSourceObjectNumbers: ctxSource: xRefTableSize:%d trailer.Size:%d - %s\n", len(ctxSource.Table), *ctxSource.Size, ctxSource.Read.FileName)
	logInfoMerge.Printf("patchSourceObjectNumbers:   ctxDest: xRefTableSize:%d trailer.Size:%d - %s\n", len(ctxDest.Table), *ctxDest.Size, ctxDest.Read.FileName)

	// Patch source xref tables obj numbers which are essentially the keys.
	//logInfoMerge.Printf("Source XRefTable before:\n%s\n", ctxSource)

	keys := sortedKeys(ctxSource)

	// Create lookup table for obj numbers.
	lookup := lookupTable(keys, *ctxDest.Size)

	// Patch pointer to root object
	patchIndRef(ctxSource.Root, lookup)

	// Patch pointer to info object
	if ctxSource.Info != nil {
		patchIndRef(ctxSource.Info, lookup)
	}

	// Patch free object zero
	entry := ctxSource.Table[0]
	off := int(*entry.Offset)
	if off != 0 {
		i := int64(lookup[off])
		entry.Offset = &i
	}

	// Patch all indRefs for xref table entries.
	for _, k := range keys {

		//logDebugMerge.Printf("patching obj #%d\n", k)

		entry := ctxSource.Table[k]

		if entry.Free {
			logDebugMerge.Printf("patch free entry: old offset:%d\n", *entry.Offset)
			off := int(*entry.Offset)
			if off == 0 {
				continue
			}
			i := int64(lookup[off])
			entry.Offset = &i
			logDebugMerge.Printf("patch free entry: new offset:%d\n", *entry.Offset)
			continue
		}

		patchObject(entry.Object, lookup)
	}

	// Patch xref entry object numbers.
	m := make(map[int]*types.XRefTableEntry, *ctxSource.Size)
	for k, v := range lookup {
		m[v] = ctxSource.Table[k]
	}
	m[0] = ctxSource.Table[0]
	ctxSource.Table = m

	// Patch DuplicateInfo object numbers.
	ctxSource.Optimize.DuplicateInfoObjects = patchObjects(ctxSource.Optimize.DuplicateInfoObjects, lookup)

	// Patch Linearization object numbers.
	ctxSource.LinearizationObjs = patchObjects(ctxSource.LinearizationObjs, lookup)

	// Patch XRefStream objects numbers.
	ctxSource.Read.XRefStreams = patchObjects(ctxSource.Read.XRefStreams, lookup)

	// Patch object stream object numbers.
	ctxSource.Read.ObjectStreams = patchObjects(ctxSource.Read.ObjectStreams, lookup)

	logInfoMerge.Printf("patchSourceObjectNumbers end")
}

func appendSourcePageTreeToDestPageTree(ctxSource, ctxDest *types.PDFContext) error {

	logDebugMerge.Println("appendSourcePageTreeToDestPageTree begin")

	indRefPageTreeRootDictSource, err := ctxSource.Pages()
	if err != nil {
		return err
	}

	pageTreeRootDictSource, _ := ctxSource.XRefTable.DereferenceDict(*indRefPageTreeRootDictSource)
	pageCountSource := pageTreeRootDictSource.IntEntry("Count")

	indRefPageTreeRootDictDest, err := ctxDest.Pages()
	if err != nil {
		return err
	}

	pageTreeRootDictDest, _ := ctxDest.XRefTable.DereferenceDict(*indRefPageTreeRootDictDest)
	pageCountDest := pageTreeRootDictDest.IntEntry("Count")

	arr := pageTreeRootDictDest.PDFArrayEntry("Kids")
	logDebugMerge.Printf("Kids before: %v\n", *arr)

	pageTreeRootDictSource.Insert("Parent", *indRefPageTreeRootDictDest)

	// The source page tree gets appended on to the dest page tree.
	*arr = append(*arr, *indRefPageTreeRootDictSource)
	logDebugMerge.Printf("Kids after: %v\n", *arr)

	pageTreeRootDictDest.Update("Count", types.PDFInteger(*pageCountDest+*pageCountSource))
	pageTreeRootDictDest.Update("Kids", *arr)

	ctxDest.PageCount += ctxSource.PageCount

	logDebugMerge.Println("appendSourcePageTreeToDestPageTree end")

	return nil
}

func appendSourceObjectsToDest(ctxSource, ctxDest *types.PDFContext) {

	logDebugMerge.Println("appendSourceObjectsToDest begin")

	for objNr, entry := range ctxSource.Table {

		// Do not copy free list head.
		if objNr == 0 {
			continue
		}

		logDebugMerge.Printf("adding obj %d from src to dest\n", objNr)

		ctxDest.Table[objNr] = entry

		*ctxDest.Size++

	}

	logDebugMerge.Println("appendSourceObjectsToDest end")
}

// merge two disjunct IntSets
func mergeIntSets(src, dest types.IntSet) {
	for k := range src {
		dest[k] = true
	}
}

func mergeDuplicateObjNumberIntSets(ctxSource, ctxDest *types.PDFContext) {

	logDebugMerge.Println("mergeDuplicateObjNumberIntSets begin")

	mergeIntSets(ctxSource.Optimize.DuplicateInfoObjects, ctxDest.Optimize.DuplicateInfoObjects)
	mergeIntSets(ctxSource.LinearizationObjs, ctxDest.LinearizationObjs)
	mergeIntSets(ctxSource.Read.XRefStreams, ctxDest.Read.XRefStreams)
	mergeIntSets(ctxSource.Read.ObjectStreams, ctxDest.Read.ObjectStreams)

	logDebugMerge.Println("mergeDuplicateObjNumberIntSets end")
}

// XRefTables merges PDFContext ctxSource into ctxDest by appending its page tree.
func XRefTables(ctxSource, ctxDest *types.PDFContext) (err error) {

	// Sweep over ctxSource cross ref table and ensure valid object numbers in ctxDest's space.
	patchSourceObjectNumbers(ctxSource, ctxDest)

	// Append ctxSource pageTree to ctxDest pageTree.
	logInfoMerge.Println("appendSourcePageTreeToDestPageTree")
	err = appendSourcePageTreeToDestPageTree(ctxSource, ctxDest)
	if err != nil {
		return err
	}

	// Append ctxSource objects to ctxDest
	logInfoMerge.Println("appendSourceObjectsToDest")
	appendSourceObjectsToDest(ctxSource, ctxDest)

	// Mark source's root object as free.
	err = ctxDest.DeleteObject(int(ctxSource.Root.ObjectNumber))
	if err != nil {
		return
	}

	// Mark source's info object as free.
	// Note: Any indRefs this info object depends on are missed.
	if ctxSource.Info != nil {
		err = ctxDest.DeleteObject(int(ctxSource.Info.ObjectNumber))
		if err != nil {
			return
		}
	}

	// Merge all IntSets containing redundant object numbers.
	logInfoMerge.Println("mergeDuplicateObjNumberIntSets")
	mergeDuplicateObjNumberIntSets(ctxSource, ctxDest)

	logInfoMerge.Printf("Dest XRefTable after merge:\n%s\n", ctxDest)

	return nil
}
