// Tideland Go Library - Numerics
//
// Copyright (C) 2009-2015 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package numerics

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"math"
	"sort"
)

//--------------------
// POINT
//--------------------

// Point is just one point in a 2D coordinate system. The
// values for x or x are read-only.
type Point struct {
	x float64
	y float64
}

// NewPoint creates a new point.
func NewPoint(x, y float64) *Point {
	return &Point{x, y}
}

// IsInf checks if x or y is infinite.
func (p Point) IsInf() bool {
	return math.IsInf(p.x, 0) || math.IsInf(p.y, 0)
}

// IsNaN checks if x or y is not a number.
func (p Point) IsNaN() bool {
	return math.IsNaN(p.x) || math.IsNaN(p.y)
}

// X returns the x value of the point.
func (p Point) X() float64 {
	return p.x
}

// Y returns the y value of the point.
func (p Point) Y() float64 {
	return p.y
}

// DistanceTo takes another point and calculates the
// geometric distance.
func (p Point) DistanceTo(op *Point) float64 {
	dx := p.x - op.x
	dy := p.y - op.y
	return math.Sqrt(dx*dx + dy*dy)
}

// VectorTo returns the vector to another point.
func (p Point) VectorTo(op *Point) *Vector {
	return NewVector(op.X()-p.x, op.Y()-p.y)
}

// String returns the string representation of the coordinates.
func (p Point) String() string {
	return fmt.Sprintf("(%f, %f)", p.x, p.y)
}

// MiddlePoint returns the middle point between two points.
func MiddlePoint(a, b *Point) *Point {
	return NewPoint((a.x+b.x)/2, (a.y+b.y)/2)
}

// PointVector returns the vector between two poins.
func PointVector(a, b *Point) *Vector {
	return a.VectorTo(b)
}

//--------------------
// POINTS
//--------------------

// Points is just a set of points.
type Points []*Point

// NewPoints creates a set of points.
func NewPoints(p ...*Point) Points {
	if len(p) > 0 {
		return p
	}
	return []*Point{}
}

// XAt returns the X value of the point at a given index.
func (ps Points) XAt(idx int) float64 {
	return ps[idx].X()
}

// YAt returns the Y value of the point at a given index.
func (ps Points) YAt(idx int) float64 {
	return ps[idx].Y()
}

// XDifference returns the difference between two X
// values of the set.
func (ps Points) XDifference(idxA, idxB int) float64 {
	return ps[idxA].X() - ps[idxB].X()
}

// YDifference returns the difference between two Y
// values of the set.
func (ps Points) YDifference(idxA, idxB int) float64 {
	return ps[idxA].Y() - ps[idxB].Y()
}

// XInRange tests if an X value is in the range of X
// values of the set.
func (ps Points) XInRange(x float64) bool {
	minX := ps[0].X()
	maxX := ps[0].X()
	for _, p := range ps[1:] {
		if p.X() < minX {
			minX = p.X()
		}
		if p.X() > maxX {
			maxX = p.X()
		}
	}
	return minX <= x && x <= maxX
}

// SearchNextIndex searches the next index fo a
// given X value.
func (ps Points) SearchNextIndex(x float64) int {
	sf := func(i int) bool {
		return x < ps[i].X()
	}
	return sort.Search(len(ps), sf)
}

// Len returns the number of points in the set.
func (ps Points) Len() int {
	return len(ps)
}

// Less returns true if the point with index i is less then the
// one with index j. It first looks for X, then for Y.
func (ps Points) Less(i, j int) bool {
	// Check X first.
	switch {
	case ps[i].x < ps[j].x:
		return true
	case ps[i].x > ps[j].x:
		return false
	}
	// Now check Y.
	switch {
	case ps[i].y < ps[j].y:
		return true
	case ps[i].y > ps[j].y:
		return false
	}
	return false
}

// Swap swaps two points of the set.
func (ps Points) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}

// CubicSplineFunction returns a cubic spline function based on the points.
func (ps Points) CubicSplineFunction() *CubicSplineFunction {
	return NewCubicSplineFunction(ps)
}

// LeastSquaresFunction returns a least squares function based on the points.
func (ps Points) LeastSquaresFunction() *LeastSquaresFunction {
	return NewLeastSquaresFunction(ps)
}

// String returns the string representation of the set.
func (ps Points) String() string {
	pss := "{"
	for _, p := range ps {
		pss += p.String()
	}
	pss += "}"
	return pss
}

//--------------------
// VECTOR
//--------------------

// Vector represents a vector in a coordinate system. The
// values are read-only.
type Vector struct {
	x float64
	y float64
}

// NewVector creates a new vector.
func NewVector(x, y float64) *Vector {
	return &Vector{x, y}
}

// X returns the x value of the vector.
func (v Vector) X() float64 {
	return v.x
}

// Y returns the y value of the vector.
func (v Vector) Y() float64 {
	return v.y
}

// Len returns the length of the vector.
func (v Vector) Len() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y)
}

// String returns the string representation of the vector.
func (v Vector) String() string {
	return fmt.Sprintf("<%f, %f>", v.x, v.y)
}

// AddVectors returns a new vector as addition of two vectors.
func AddVectors(a, b *Vector) *Vector {
	return NewVector(a.x+b.x, a.y+b.y)
}

// SubVectors returns a new vector as subtraction of two vectors.
func SubVectors(a, b *Vector) *Vector {
	return NewVector(a.x-b.x, a.y-b.y)
}

// ScaleVectors multiplies a vector with a float and returns
// the new vector.
func ScaleVector(v *Vector, s float64) *Vector {
	return NewVector(v.x*s, v.y*s)
}

//--------------------
// FUNCTION
//--------------------

// Function is the standard interface the nmerical
// functions have to implement.
type Function interface {
	// Eval evaluates a function for the value x.
	Eval(x float64) float64

	// EvalPoint evaluates a function for the value
	// x and returns the result as point.
	EvalPoint(x float64) *Point

	// EvalPoints evaluates the function count times
	// with values between fromX and toX. The result is
	// returned as a set of pints.
	EvalPoints(fromX, toX float64, count int) Points
}

//--------------------
// POLYNOMIAL FUNCTION
//--------------------

// PolynomialFunction is a polynomial function based on a number
// of coefficients.
type PolynomialFunction struct {
	coefficients []float64
}

// NewPolynomialFunction creates a new polynomial function.
func NewPolynomialFunction(coefficients []float64) *PolynomialFunction {
	if len(coefficients) < 1 {
		return nil
	}
	pf := &PolynomialFunction{
		coefficients: coefficients,
	}
	return pf
}

// Eval evaluates the function for a given X value and
// returns the Y value.
func (pf PolynomialFunction) Eval(x float64) float64 {
	n := len(pf.coefficients)
	result := pf.coefficients[n-1]
	for i := n - 2; i >= 0; i-- {
		result = x*result + pf.coefficients[i]
	}
	return result
}

// EvalPoint evaluates the function for a given X value
// and returns the result as a point.
func (pf PolynomialFunction) EvalPoint(x float64) *Point {
	return NewPoint(x, pf.Eval(x))
}

// EvalPoints evaluates the function for a range of X values
// and returns the result as a set of points.
func (pf PolynomialFunction) EvalPoints(fromX, toX float64, count int) Points {
	return evalPoints(pf, fromX, toX, count)
}

// Differentiate differentiates the polynomial and returns the
// new polynomial.
func (pf PolynomialFunction) Differentiate() *PolynomialFunction {
	n := len(pf.coefficients)
	if n == 1 {
		return NewPolynomialFunction([]float64{0.0})
	}
	newCoefficients := make([]float64, n-1)
	for i := n - 1; i > 0; i-- {
		newCoefficients[i-1] = float64(i) * pf.coefficients[i]
	}
	return NewPolynomialFunction(newCoefficients)
}

// String returns the string representation of the function
// as f(x) := 2.9x^3+x^2-3.3x+1.0.
func (pf PolynomialFunction) String() string {
	pfs := "f(x) := "
	for i := len(pf.coefficients) - 1; i > 0; i-- {
		if pf.coefficients[i] != 0.0 {
			pfs += fmt.Sprintf("%vx", pf.coefficients[i])
			if i > 1 {
				pfs += fmt.Sprintf("^%v", i)
			}
			if pf.coefficients[i-1] > 0 {
				pfs += "+"
			}
		}
	}
	if pf.coefficients[0] != 0.0 {
		pfs += fmt.Sprintf("%v", pf.coefficients[0])
	}
	return pfs
}

//--------------------
// CUBIC SPLINE FUNCTION
//--------------------

// CubicSplineFunction is a function based on polynamial functions
// and a set of points it is going through.
type CubicSplineFunction struct {
	polynomials []*PolynomialFunction
	points      Points
}

// NewCubicSplineFunction creates a cubic spline function based on a
// set of points.
func NewCubicSplineFunction(points Points) *CubicSplineFunction {
	if points.Len() < 3 {
		return nil
	}
	csf := &CubicSplineFunction{
		points: points,
	}
	// Calculating differences between points.
	intervals := points.Len() - 1
	differences := make([]float64, intervals)
	for i := 0; i < intervals; i++ {
		differences[i] = points[i+1].X() - points[i].X()
	}
	mu := make([]float64, intervals)
	z := make([]float64, points.Len())
	var g float64
	for i := 1; i < intervals; i++ {
		g = 2.0*points.XDifference(i+1, i-1) - differences[i-1]*mu[i-1]
		mu[i] = differences[i] / g
		z[i] = (3.0*(points.YAt(i+1)*differences[i-1]-points.YAt(i)*
			points.XDifference(i+1, i-1)+points.YAt(i-1)*differences[i])/
			(differences[i-1]*differences[i]) - differences[i-1]*z[i-1]) / g
	}
	// Cubic spline coefficients (b is linear, c is quadratic, d is cubic).
	b := make([]float64, intervals)
	c := make([]float64, points.Len())
	d := make([]float64, intervals)
	for i := intervals - 1; i >= 0; i-- {
		c[i] = z[i] - mu[i]*c[i+1]
		b[i] = points.YDifference(i+1, i)/differences[i] - differences[i]*(c[i+1]+2.0*c[i])/3.0
		d[i] = (c[i+1] - c[i]) / (3.0 * differences[i])
	}
	// Build polymonials.
	csf.polynomials = make([]*PolynomialFunction, intervals)
	coefficients := make([]float64, 4)
	for i := 0; i < intervals; i++ {
		coefficients[0] = points.YAt(i)
		coefficients[1] = b[i]
		coefficients[2] = c[i]
		coefficients[3] = d[i]
		csf.polynomials[i] = NewPolynomialFunction(coefficients)
	}
	return csf
}

// Eval evaluates the function for a given X value and
// returns the Y value.
func (csf *CubicSplineFunction) Eval(x float64) float64 {
	if !csf.points.XInRange(x) {
		panic("X out of range!")
	}
	idx := csf.points.SearchNextIndex(x)
	if idx >= len(csf.polynomials) {
		idx = len(csf.polynomials) - 1
	}
	return csf.polynomials[idx].Eval(x - csf.points.XAt(idx))
}

// EvalPoint evaluates the function for a given X value
// and returns the result as a point.
func (csf *CubicSplineFunction) EvalPoint(x float64) *Point {
	return NewPoint(x, csf.Eval(x))
}

// EvalPoints evaluates the function for a range of X values
// and returns the result as a set of points.
func (csf *CubicSplineFunction) EvalPoints(fromX, toX float64, count int) Points {
	return evalPoints(csf, fromX, toX, count)
}

//--------------------
// LEAST SQUARES FUNCTION
//--------------------

// LeastSquaresFunction is a function for approximation.
type LeastSquaresFunction struct {
	sumX, sumXX float64
	sumY, sumYY float64
	sumXY       float64
	xBar, yBar  float64
	count       int
}

// NewLeastSquaresFunction creates a new least squares function based
// on a set of points.
func NewLeastSquaresFunction(points Points) *LeastSquaresFunction {
	lsf := new(LeastSquaresFunction)
	if points != nil {
		lsf.AppendPoints(points)
	}
	return lsf
}

// AppendPoint appends one point to the function.
func (lsf *LeastSquaresFunction) AppendPoint(x, y float64) {
	p := NewPoint(x, y)
	if lsf.count == 0 {
		lsf.xBar = p.X()
		lsf.yBar = p.Y()
	} else {
		dx := p.X() - lsf.xBar
		dy := p.Y() - lsf.yBar

		lsf.sumXX += dx * dx * float64(lsf.count) / float64(lsf.count+1.0)
		lsf.sumYY += dy * dy * float64(lsf.count) / float64(lsf.count+1.0)
		lsf.sumXY += dx * dy * float64(lsf.count) / float64(lsf.count+1.0)

		lsf.xBar += dx / float64(lsf.count+1.0)
		lsf.yBar += dy / float64(lsf.count+1.0)
	}
	lsf.sumX += p.X()
	lsf.sumY += p.Y()
	lsf.count++
}

// AppendPoints appends a set of points to the function.
func (lsf *LeastSquaresFunction) AppendPoints(points Points) {
	for _, p := range points {
		lsf.AppendPoint(p.X(), p.Y())
	}
}

// Eval evaluates the function for a given X value and
// returns the Y value.
func (lsf *LeastSquaresFunction) Eval(x float64) float64 {
	slope := lsf.slope()
	result := lsf.intercept(slope) + slope*x
	return result
}

// EvalPoint evaluates the function for a given X value
// and returns the result as a point.
func (lsf *LeastSquaresFunction) EvalPoint(x float64) *Point {
	return NewPoint(x, lsf.Eval(x))
}

// EvalPoints evaluates the function for a range of X values
// and returns the result as a set of points.
func (lsf *LeastSquaresFunction) EvalPoints(fromX, toX float64, count int) Points {
	return evalPoints(lsf, fromX, toX, count)
}

// slope returns the slope of the least square function.
func (lsf *LeastSquaresFunction) slope() float64 {
	if lsf.count < 2 {
		// Not enough points added.
		return math.NaN()
	}

	if math.Abs(lsf.sumXX) < 10*math.SmallestNonzeroFloat64 {
		// Not enough variation in X.
		return math.NaN()
	}
	return lsf.sumXY / lsf.sumXX
}

// intercept returns the intercept for a given slope.
func (lsf *LeastSquaresFunction) intercept(slope float64) float64 {
	return (lsf.sumY - slope*lsf.sumX) / float64(lsf.count)
}

//--------------------
// HELPERS
//--------------------

// evalPoints evaluate a function for a range and a
// number of evaluations.
func evalPoints(f Function, fromX, toX float64, count int) Points {
	interval := (toX - fromX) / float64(count)
	ps := NewPoints()
	for x := fromX; x < toX; x += interval {
		y := f.Eval(x)
		ps = append(ps, NewPoint(x, y))
	}
	return ps
}

// EOF
