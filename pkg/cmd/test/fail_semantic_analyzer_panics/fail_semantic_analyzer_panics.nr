package fail_semantic_analyzer_panics

pub type Result[T, E] = sum {
    Ok: { val: T },
    Err: { err: E }
}

pub fn (self: &Result[T, E]) IsErr[T, E]() bool {
    return false
}

pub fn GetResult[T]() Result[T, str] {
    var res: Result[T, str]
    return @res
}

pub fn main() i32 {
    // Type argument T is missing and cannot be inferred.
    var parse_res = GetResult()
    
    // Calling IsErr() on it will fail to infer type arguments, 
    // leading to `sa.SemanticInfo.Types[n]` being implicitly nil 
    // before the compiler patch.
    if parse_res.IsErr() {
        return 1
    }
    
    return 0
}
