package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

var inventory_start_regexp = regexp.MustCompile(`\x00InventoryModel\x00`)
var inventory_end_regexp = regexp.MustCompile(`\x00MarketModel\x00`)
var roster_start_regexp = regexp.MustCompile(`\x00RosterModel\x00`)
var roster_end_regexp = regexp.MustCompile(`\x00ToiModel\x00`)
var save_start_regexp = regexp.MustCompile(`\x00SaveStateModel\x00`)
var save_end_regexp = regexp.MustCompile(`\x00DataCacheModel\x00`)
var reputation_regexp = regexp.MustCompile(`\x00Reputation\x00.{33}`)

func usage() {
	fmt.Fprintln(os.Stderr, "Mech Warrior 5 New Game Creator - Resets campaign progress for a mech warrior 5 mercenaries save file.")
	fmt.Fprintln(os.Stderr, "<https://github.com/KarmaPenny/MW5NG>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage: mw5ng [FILE]")
	fmt.Fprintln(os.Stderr, "Example: mw5ng file.sav")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Copyright (C) 2019 Cole Robinette")
	fmt.Fprintln(os.Stderr, "This program is free to use, redistribute, and modify under")
	fmt.Fprintln(os.Stderr, "the terms of the GNU General Public License version 3. This")
	fmt.Fprintln(os.Stderr, "program is distributed without any warranty.")
	fmt.Fprintln(os.Stderr, "<https://www.gnu.org/licenses/>")
}

func init() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
		os.Exit(1)
	}
}

func main() {
	// read save data into buffer
	data, err := ioutil.ReadFile(flag.Arg(0))
	if (err != nil) {
		panic(err)
	}

	// read fresh game save data into buffer
	ex, err := os.Executable()
	if err != nil {
			panic(err)
	}
	exPath := filepath.Dir(ex)
	fresh_path := path.Join(exPath, "fresh_game.sav")
	new_data, err := ioutil.ReadFile(fresh_path)
	if (err != nil) {
		panic(err)
	}

	// copy xp data
	new_data = reputation_regexp.ReplaceAllLiteral(new_data, reputation_regexp.Find(data))

	// create new save buffer
	var f bytes.Buffer

	// copy initial new data
	end := inventory_start_regexp.FindIndex(new_data)
	f.Write(new_data[0:end[0]])

	// copy inventory
	start := inventory_start_regexp.FindIndex(data)
	end = inventory_end_regexp.FindIndex(data)
	f.Write(data[start[0]:end[0]])

	// copy filler from new data
	start = inventory_end_regexp.FindIndex(new_data)
	end = roster_start_regexp.FindIndex(new_data)
	f.Write(new_data[start[0]:end[0]])

	// copy roster
	start = roster_start_regexp.FindIndex(data)
	end = roster_end_regexp.FindIndex(data)
	f.Write(data[start[0]:end[0]])

	// copy filler from new data
	start = roster_end_regexp.FindIndex(new_data)
	end = save_start_regexp.FindIndex(new_data)
	f.Write(new_data[start[0]:end[0]])

	// copy save state
	start = save_start_regexp.FindIndex(data)
	end = save_end_regexp.FindIndex(data)
	f.Write(data[start[0]:end[0]])

	// copy remaining filler from new data
	start = save_end_regexp.FindIndex(new_data)
	f.Write(new_data[start[0]:])

	// get byte array of new save data
	fbytes := f.Bytes()

	// fix Persistent data length properties
	ap := len(fbytes) - 187
	bp := len(fbytes) - 191
	apb := make([]byte, 4)
	binary.LittleEndian.PutUint32(apb, uint32(ap))
	bpb := make([]byte, 4)
	binary.LittleEndian.PutUint32(bpb, uint32(bp))
	for i := 0; i < 4; i++ {
		fbytes[i+148] = apb[i]
		fbytes[i+174] = bpb[i]
	}

	// write save data to file
	err = ioutil.WriteFile(flag.Arg(0), fbytes, 0644)
}
