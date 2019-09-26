package models

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestSourceHeader1(t *testing.T) {
	filePath := os.TempDir() + "testSourceHeader.file"
	expected := `//header`
	err := ioutil.WriteFile(filePath, []byte(expected), os.ModePerm)
	if err != nil {
		println(filePath)
		println(err)
		t.Fail()
	}
	defer func() {
		_ = os.Remove(filePath)
	}()
	source := Source{Path: filePath}
	if source.Header() != expected {
		t.Fail()
	}
}

func TestSourceHeader2(t *testing.T) {
	filePath := os.TempDir() + "testSourceHeader.file"
	expected := `/*header*/`
	err := ioutil.WriteFile(filePath, []byte(expected), os.ModePerm)
	if err != nil {
		println(filePath)
		println(err)
		t.Fail()
	}
	defer func() {
		_ = os.Remove(filePath)
	}()
	source := Source{Path: filePath}
	if source.Header() != expected {
		t.Fail()
	}
}
func TestSourceHeader3(t *testing.T) {
	filePath := os.TempDir() + "testSourceHeader.file"
	expected := `/*
	header
	*/`
	err := ioutil.WriteFile(filePath, []byte(expected), os.ModePerm)
	if err != nil {
		println(filePath)
		println(err)
		t.Fail()
	}
	defer func() {
		_ = os.Remove(filePath)
	}()
	source := Source{Path: filePath}
	if source.Header() != expected {
		println(source.Header())
		t.Fail()
	}
}
func TestSourceHeader4(t *testing.T) {
	filePath := os.TempDir() + "testSourceHeader.file"
	expected := `//
//`
	err := ioutil.WriteFile(filePath, []byte(expected), os.ModePerm)
	if err != nil {
		println(filePath)
		println(err)
		t.Fail()
	}
	defer func() {
		_ = os.Remove(filePath)
	}()
	source := Source{Path: filePath}
	if source.Header() != expected {
		t.Fail()
	}
}
