package math

// FFI Declarations mapping directly to standard C <math.h> routines.
// In Nora, FFI functions marked as 'extern fn' are kept unmangled by the transpiler.
extern fn sin(x: f64) f64

extern fn cos(x: f64) f64

extern fn tan(x: f64) f64

extern fn asin(x: f64) f64

extern fn acos(x: f64) f64

extern fn atan(x: f64) f64

extern fn atan2(y: f64, x: f64) f64

extern fn sinh(x: f64) f64

extern fn cosh(x: f64) f64

extern fn tanh(x: f64) f64

extern fn pow(x: f64, y: f64) f64

extern fn sqrt(x: f64) f64

extern fn cbrt(x: f64) f64

extern fn exp(x: f64) f64

extern fn log(x: f64) f64

extern fn log2(x: f64) f64

extern fn log10(x: f64) f64

extern fn floor(x: f64) f64

extern fn ceil(x: f64) f64

extern fn round(x: f64) f64

extern fn trunc(x: f64) f64

extern fn fabs(x: f64) f64

extern fn fmax(x: f64, y: f64) f64

extern fn fmin(x: f64, y: f64) f64

extern fn fmod(x: f64, y: f64) f64

extern fn remainder(x: f64, y: f64) f64

// --- Math Constants ---
pub fn Pi() f64 {
    return 3.141592653589793
}

pub fn E() f64 {
    return 2.718281828459045
}

pub fn Tau() f64 {
    return 6.283185307179586
}

pub fn Sqrt2() f64 {
    return 1.414213562373095
}

pub fn Ln2() f64 {
    return 0.6931471805599453
}

pub fn Ln10() f64 {
    return 2.302585092994046
}

// --- Trigonometry ---
pub fn Sin(x: f64) f64 {
    return sin(x)
}

pub fn Cos(x: f64) f64 {
    return cos(x)
}

pub fn Tan(x: f64) f64 {
    return tan(x)
}

pub fn Asin(x: f64) f64 {
    return asin(x)
}

pub fn Acos(x: f64) f64 {
    return acos(x)
}

pub fn Atan(x: f64) f64 {
    return atan(x)
}

pub fn Atan2(y: f64, x: f64) f64 {
    return atan2(y, x)
}

pub fn Sinh(x: f64) f64 {
    return sinh(x)
}

pub fn Cosh(x: f64) f64 {
    return cosh(x)
}

pub fn Tanh(x: f64) f64 {
    return tanh(x)
}

// --- Powers & Logarithms ---
pub fn Pow(x: f64, y: f64) f64 {
    return pow(x, y)
}

pub fn Sqrt(x: f64) f64 {
    return sqrt(x)
}

pub fn Cbrt(x: f64) f64 {
    return cbrt(x)
}

pub fn Exp(x: f64) f64 {
    return exp(x)
}

pub fn Log(x: f64) f64 {
    return log(x)
}

pub fn Log2(x: f64) f64 {
    return log2(x)
}

pub fn Log10(x: f64) f64 {
    return log10(x)
}

pub fn Ln(x: f64) f64 {
    return log(x)
}

// --- Rounding & Floating-Point Arithmetic ---
pub fn Floor(x: f64) f64 {
    return floor(x)
}

pub fn Ceil(x: f64) f64 {
    return ceil(x)
}

pub fn Round(x: f64) f64 {
    return round(x)
}

pub fn Trunc(x: f64) f64 {
    return trunc(x)
}

pub fn Abs(x: f64) f64 {
    return fabs(x)
}

pub fn Max(x: f64, y: f64) f64 {
    return fmax(x, y)
}

pub fn Min(x: f64, y: f64) f64 {
    return fmin(x, y)
}

pub fn Mod(x: f64, y: f64) f64 {
    return fmod(x, y)
}

pub fn Remainder(x: f64, y: f64) f64 {
    return remainder(x, y)
}
