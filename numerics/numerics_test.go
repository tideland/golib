// Tideland Go Library - Numerics - Unit Tests
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
	"sort"
	"testing"

	"github.com/tideland/golib/audit"
)

//--------------------
// TESTS
//--------------------

// Test simple point.
func TestSimplePoint(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Create some points.
	pa := NewPoint(1.0, 5.0)
	pb := NewPoint(2.0, 2.0)
	pc := NewPoint(3.0, 4.0)
	// Asserts.
	assert.Equal(pa.X(), 1.0, "X of point A")
	assert.Equal(pa.Y(), 5.0, "Y of point A")
	assert.About(pb.DistanceTo(pc), 2.236, 0.0001, "distance B to C")
	assert.Equal(MiddlePoint(pb, pc).X(), 2.5, "X value of middle point")
	assert.Equal(MiddlePoint(pb, pc).Y(), 3.0, "Y value of middle point")
	assert.Equal(PointVector(pb, pc).String(), "<1.000000, 2.000000>", "string representation of vector")
}

// Test simple point array.
func TestSimplePointArray(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Create some points.
	ps := NewPoints()
	assert.Empty(ps, "No points yet.")
	ps = append(ps, NewPoint(2.0, 2.0))
	ps = append(ps, NewPoint(5.0, 1.0))
	ps = append(ps, NewPoint(4.0, 2.0))
	ps = append(ps, NewPoint(3.0, 3.0))
	ps = append(ps, NewPoint(1.0, 1.0))
	// Asserts.
	assert.Equal(ps.Len(), 5, "now with points")
	sort.Sort(ps)
	assert.Equal(ps[0], NewPoint(1.0, 1.0), "first point")
	assert.Equal(ps.Len(), 5, "length")
	assert.Equal(ps[3].X(), 4.0, "X of 4th point")
}

// Test simple polynomial function.
func TestSimplePolynomialFunction(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Some polynominal function evaluations.
	p := NewPolynomialFunction([]float64{2.0, 2.0})
	fa := p.Eval(-2.0)
	fb := p.Eval(-1.0)
	fc := p.Eval(0.0)
	fd := p.Eval(2.0)
	// Asserts.
	assert.Equal(fa, -2.0, "f(a)")
	assert.Equal(fb, 0.0, "f(b)")
	assert.Equal(fc, 2.0, "f(c)")
	assert.Equal(fd, 6.0, "f(d)")
}

// Test polynomial function printing.
func TestPolynomialFunctionPrinting(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	p := NewPolynomialFunction([]float64{-7.55, 2.0, -3.1, 2.66, -3.45})
	// Asserts.
	assert.Equal(p.String(), "f(x) := -3.45x^4+2.66x^3-3.1x^2+2x-7.55", "string representation")
}

// Test quadratic polynomial function.
func TestQuadraticPolynomialFunction(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Some polynominal function evaluations.
	p := NewPolynomialFunction([]float64{0.0, 0.0, 1.0})
	fa := p.Eval(-1.0)
	fb := p.Eval(2.0)
	fc := p.Eval(-3.0)
	// Asserts.
	assert.Equal(fa, 1.0, "f(a)")
	assert.Equal(fb, 4.0, "f(b)")
	assert.Equal(fc, 9.0, "f(c)")
}

// Test polynomial function differentiation.
func TestPolynomialFunctionDifferentiation(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	p := NewPolynomialFunction([]float64{1.0, 2.0, 1.0, 3.0})
	dp := p.Differentiate()
	// Asserts.
	assert.Equal(dp.String(), "f(x) := 9x^2+2x+2")
}

// Test interpolation.
func TestInterpolation(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Build a cubic spline function.
	ps := NewPoints()
	ps = append(ps, NewPoint(1.0, 1.0))
	ps = append(ps, NewPoint(2.0, 2.0))
	ps = append(ps, NewPoint(3.0, 3.0))
	ps = append(ps, NewPoint(4.0, 2.0))
	ps = append(ps, NewPoint(5.0, 1.0))
	f := NewCubicSplineFunction(ps)
	// Asserts.
	assert.About(f.EvalPoint(3.5).Y(), 2.7678, 0.0001, "f(3.5)")
}

// Test points evaluation.
func TestPointsEvaluation(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Build a least squares function out of the results
	// off a cubic spline function.
	ps := NewPoints()
	ps = append(ps, NewPoint(0.0, 0.7))
	ps = append(ps, NewPoint(1.0, 1.1))
	ps = append(ps, NewPoint(2.0, 0.0))
	ps = append(ps, NewPoint(3.0, -0.5))
	ps = append(ps, NewPoint(4.0, -2.0))
	ps = append(ps, NewPoint(5.0, -1.0))
	ps = append(ps, NewPoint(6.0, 0.2))
	ps = append(ps, NewPoint(7.0, 0.3))
	ps = append(ps, NewPoint(8.0, -0.4))
	ps = append(ps, NewPoint(9.0, -0.5))
	lsf := ps.CubicSplineFunction().EvalPoints(0.7, 8.1, 50).LeastSquaresFunction()
	// Asserts.
	assert.About(lsf.Eval(15.0), -0.0407, 0.0001, "f(15.0)")
}

// Test least squares function.
func TestLeastSquaresFunction(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	// Build a leas squares function.
	lsf := NewLeastSquaresFunction(nil)
	lsf.AppendPoint(1.0, 1.0)
	lsf.AppendPoint(2.0, 0.5)
	lsf.AppendPoint(3.0, 2.0)
	lsf.AppendPoint(4.0, 2.5)
	lsf.AppendPoint(5.0, 1.5)
	lsf.AppendPoint(6.0, 1.0)
	lsf.AppendPoint(7.0, 1.5)
	// Asserts.
	assert.About(lsf.Eval(9.0), 1.7857, 0.0001, "f(9.0)")
	assert.About(lsf.Eval(4.5), 1.4642, 0.0001, "f(4.5)")
}

// EOF
