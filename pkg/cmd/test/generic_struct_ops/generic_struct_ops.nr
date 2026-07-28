package main

pub type MyFixed = struct {
	val: i32
}

pub fn (f: #MyFixed) operator==(other: MyFixed) bool {
	return f.val == other.val
}

pub fn (f: #MyFixed) operator<(other: MyFixed) bool {
	return f.val < other.val
}

pub type Vector3[T] = struct {
	x: T,
	y: T,
	z: T
}

pub fn (v: #Vector3[T]) Equals[T](other: #Vector3[T]) bool {
	return v.x == other.x && v.y == other.y && v.z == other.z
}

pub fn (v: #Vector3[T]) LessThan[T](other: #Vector3[T]) bool {
	return v.x < other.x && v.y < other.y && v.z < other.z
}

pub fn main() {
	var a = Vector3[MyFixed]{x: MyFixed{val: 1}, y: MyFixed{val: 2}, z: MyFixed{val: 3}}
	var b = Vector3[MyFixed]{x: MyFixed{val: 1}, y: MyFixed{val: 2}, z: MyFixed{val: 3}}
	var c = Vector3[MyFixed]{x: MyFixed{val: 4}, y: MyFixed{val: 5}, z: MyFixed{val: 6}}

	if !a.Equals(#b) {
		panic("a should equal b")
	}

	if a.Equals(#c) {
		panic("a should not equal c")
	}

	if !a.LessThan(#c) {
		panic("a should be less than c")
	}

	if c.LessThan(#a) {
		panic("c should not be less than a")
	}
}
