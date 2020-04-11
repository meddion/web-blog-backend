package models

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func mockName() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func createMockedFile(t *testing.T) *File {
	mockedFile := NewEmptyFile("/test", mockName(), "custom")
	if _, err := mockedFile.Save(context.TODO()); err != nil {
		t.Errorf("on saving a mocking file (%v) in db: %s", mockedFile, err.Error())
	}
	return mockedFile
}

func cleanMockedFile(t *testing.T, f *File) {
	if _, err := f.Delete(context.TODO()); err != nil {
		t.Errorf("on deleting the file (%v) from db: %s", f, err.Error())
	}
}

func TestSave(t *testing.T) {
	type (
		out struct {
			modifedCount  int64
			upsertedCount int64
		}
		template struct {
			in       *File
			expected out
		}
	)
	cases := make([]template, 2)
	filename := mockName()
	cases[0] = template{NewEmptyFile("/test", filename, "custom"), out{0, 1}}
	time.Sleep(5 * time.Millisecond)
	cases[1] = template{NewEmptyFile("/test", filename, "custom"), out{1, 0}}

	for _, c := range cases {
		r, err := c.in.Save(context.TODO())
		if err != nil {
			t.Fatalf("on saving a file to the db: %s", err.Error())
		}
		defer cleanMockedFile(t, c.in)

		got := out{r.ModifiedCount, r.UpsertedCount}

		if got != c.expected {
			t.Fatalf("with input %#v the expected output had to be: %v, insted we got: %v",
				c.in, c.expected, got)
		}
	}
}

func TestGet(t *testing.T) {
	for i := 0; i < 10; i++ {
		f := createMockedFile(t)
		defer cleanMockedFile(t, f)
		if err := f.Get(context.TODO()); err != nil {
			t.Fatalf("on getting a mocked record from db: %s", err.Error())
		}
	}
}

func TestDelete(t *testing.T) {
	file := createMockedFile(t)
	cleanMockedFile(t, file)
}

func TestListFilenamesWhere(t *testing.T) {
	for i := 0; i < 10; i++ {
		f := createMockedFile(t)
		defer cleanMockedFile(t, f)
	}
	if filenames, err := ListFilenamesWhere(context.TODO(), "/test", "custom"); err != nil {
		t.Fatalf("on getting filenames slice from db: %s", err.Error())
	} else if len(filenames) != 10 {
		t.Fatal("on getting a wrong length data, expected 10 records")
	}
}
