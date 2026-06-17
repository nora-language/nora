package collections

import "io"

type ListNode[T] = struct {
    value: T,
    prev: #ListNode[T],
    next: @ListNode[T]
}

[core_intrinsic("Collection")]
pub type LinkedList[T] = struct {
    head: @ListNode[T],
    tail: #ListNode[T],
    size: i32
}

pub type LinkedListIterator[T] = struct {
    curr: #ListNode[T]
}

pub fn NewLinkedList[T]() @LinkedList[T] {
    return alloc LinkedList[T] {
        head: none,
        tail: none,
        size: 0
    }
}

[unsafe]
pub fn (self: &LinkedList[T]) PushBack[T](val: T) {
    var new_node = alloc ListNode[T] {
        value: val,
        prev: none,
        next: none
    }
    if (self.head == none) {
        self.head = @new_node
        self.tail = #self.head
    } else {
        var prev_tail = self.tail
        new_node.prev = prev_tail
        prev_tail.next = @new_node
        self.tail = #prev_tail.next
    }
    self.size = self.size + 1
}

[unsafe]
pub fn (self: &LinkedList[T]) PushFront[T](val: T) {
    var new_node = alloc ListNode[T] {
        value: val,
        prev: none,
        next: none
    }
    if (self.head == none) {
        self.head = @new_node
        self.tail = #self.head
    } else {
        var old_head = @self.head
        old_head.prev = #new_node
        new_node.next = @old_head
        self.head = @new_node
    }
    self.size = self.size + 1
}

[unsafe]
pub fn (self: &LinkedList[T]) PopFront[T]() Option[T] {
    if (self.head == none) {
        return None[T]
    }
    var old_head = @self.head
    var val = old_head.value
    if (old_head.next != none) {
        old_head.next.prev = none
        self.head = @old_head.next
    } else {
        self.tail = none
    }
    self.size = self.size - 1
    return Some[T](val)
}

[unsafe]
pub fn (self: &LinkedList[T]) PopBack[T]() Option[T] {
    if (self.head == none) {
        return None[T]
    }
    if (self.head.next == none) {
        return self.PopFront[T]()
    }
    var prev_node = self.tail.prev
    var old_tail = @prev_node.next
    self.tail = prev_node
    var val = old_tail.value
    self.size = self.size - 1
    return Some[T](val)
}

pub fn (self: &LinkedList[T]) Len[T]() i32 {
    return self.size
}

pub fn (self: &LinkedList[T]) IsEmpty[T]() bool {
    return self.size == 0
}

pub fn (self: &LinkedList[T]) Clear[T]() {
    while (self.head != none) {
        var _discarded = self.PopFront[T]()
    }
}

pub fn (self: &LinkedListIterator[T]) Next[T]() Option[#T] {
    if (self.curr == none) {
        return None[#T]
    }
    var val = #self.curr.value
    self.curr = self.curr.next
    return Some[#T](val)
}

pub fn (self: #LinkedList[T]) Iter[T]() LinkedListIterator[T] {
    return LinkedListIterator[T] {
        curr: self.head
    }
}
