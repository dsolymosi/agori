package agori

import (
	"math/rand"
	"testing"
)

func TestAgori(t *testing.T) {
	ag := NewAgoriD()
	rand.Seed(0)

	ag.InsertS("127.0.0.1")
	ag.InsertD(127, 0, 0, 1)
	ag.Insert(2130706433)
	err, b, n := ag.GetS("127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	b2, n2 := ag.GetD(127, 0, 0, 1)
	b3, n3 := ag.Get(2130706433)

	if !b {
		t.Fatal("Could not insert, or find by string")
	}
	if !b2 {
		t.Fatal("Could not insert, or find by digits")
	}
	if !b3 {
		t.Fatal("Could not insert, or find by uint32")
	}
	if n != n2 || n2 != n3 {
		t.Fatal("Count mismatch when searching for the same address using different means")
	}

	for i := 0; i < 20000; i++ {
		ag.InsertS("127.0.0.1")
	}
	err, b, n = ag.GetS("127.0.0.1")

	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatal("Could not repeatedly insert address successfully")
	}
	if n != 20003 {
		t.Fatal("Could not update value of address correctly")
	}

	for i := uint32(0); i < 30000; i++ {
		x := rand.Uint32()
		ag.Insert(x)
	}

	err, b, n = ag.GetS("127.0.0.1")

	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Fatal("Lost important address to unimportant addresses.")
	}
	if n < 20003 {
		t.Fatal("Decreased count of, or lost address to unimportant addresses.")
	}
}

func Example() {
	//create new agori with default settings
	ag := NewAgoriD()

	//insert one IP in three ways
	ag.InsertS("127.0.0.1")
	ag.InsertD(127, 0, 0, 1)
	ag.Insert(2130706433)

	//display the results
	ag.Print(0.01)
}

func BenchmarkAddExistingSingle(b *testing.B) {
	ag := NewAgoriD()
	for i := 0; i < b.N; i++ {
		ag.Insert(rand.Uint32())
	}
	ag.Insert(2130706433)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ag.Insert(2130706433)
	}
}

func BenchmarkAddExistingRandom(b *testing.B) {
	ag := NewAgori(b.N, 1.0/32.0)
	r := make([]uint32, b.N)
	for i := 0; i < b.N; i++ {
		r[i] = rand.Uint32()
		ag.Insert(r[i])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ag.Insert(r[i])
	}
}

func BenchmarkAddOrderedNewNoDelete(b *testing.B) {
	ag := NewAgori(b.N, 1.0/32.0)
	to := uint32(b.N)
	b.ResetTimer()
	for i := uint32(0); i < to; i++ {
		ag.Insert(i)
	}
}

func BenchmarkAddOrderedNew(b *testing.B) {
	ag := NewAgoriD()
	to := uint32(b.N)
	b.ResetTimer()
	for i := uint32(0); i < to; i++ {
		ag.Insert(i)
	}
}

func BenchmarkAddRandomNoDelete(b *testing.B) {
	ag := NewAgori(b.N, 1.0/32.0)
	r := make([]uint32, b.N)
	for i := 0; i < b.N; i++ {
		r[i] = rand.Uint32()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ag.Insert(r[i])
	}
}

func BenchmarkAddRandom(b *testing.B) {
	ag := NewAgoriD()
	r := make([]uint32, b.N)
	for i := 0; i < b.N; i++ {
		r[i] = rand.Uint32()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ag.Insert(r[i])
	}
}
