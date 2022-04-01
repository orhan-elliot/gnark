/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fields_bls12377

import (
	"github.com/consensys/gnark-crypto/ecc"
	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/consensys/gnark/backend/hint"
	"github.com/consensys/gnark/frontend"
	"math/big"
)

// E2 element in a quadratic extension
type E2 struct {
	A0, A1 frontend.Variable
}

// SetZero returns a newly allocated element equal to 0
func (e *E2) SetZero() *E2 {
	e.A0 = 0
	e.A1 = 0
	return e
}

// SetOne returns a newly allocated element equal to 1
func (e *E2) SetOne() *E2 {
	e.A0 = 1
	e.A1 = 0
	return e
}

func (e *E2) assign(e1 []frontend.Variable) {
	e.A0 = e1[0]
	e.A1 = e1[1]
}

// Neg negates a e2 elmt
func (e *E2) Neg(api frontend.API, e1 E2) *E2 {
	e.A0 = api.Sub(0, e1.A0)
	e.A1 = api.Sub(0, e1.A1)
	return e
}

// Add e2 elmts
func (e *E2) Add(api frontend.API, e1, e2 E2) *E2 {
	e.A0 = api.Add(e1.A0, e2.A0)
	e.A1 = api.Add(e1.A1, e2.A1)
	return e
}

// Double e2 elmt
func (e *E2) Double(api frontend.API, e1 E2) *E2 {
	e.A0 = api.Add(e1.A0, e1.A0)
	e.A1 = api.Add(e1.A1, e1.A1)
	return e
}

// Sub e2 elmts
func (e *E2) Sub(api frontend.API, e1, e2 E2) *E2 {
	e.A0 = api.Sub(e1.A0, e2.A0)
	e.A1 = api.Sub(e1.A1, e2.A1)
	return e
}

// Mul e2 elmts
func (e *E2) Mul(api frontend.API, e1, e2 E2) *E2 {

	// 1C
	l1 := api.Add(e1.A0, e1.A1)
	l2 := api.Add(e2.A0, e2.A1)

	u := api.Mul(l1, l2)

	// 2C
	ac := api.Mul(e1.A0, e2.A0)
	bd := api.Mul(e1.A1, e2.A1)

	l31 := api.Add(ac, bd)
	e.A1 = api.Sub(u, l31)

	l41 := api.Mul(bd, ext.uSquare)
	e.A0 = api.Add(ac, l41)

	return e
}

// Square e2 elt
func (e *E2) Square(api frontend.API, x E2) *E2 {
	//algo 22 https://eprint.iacr.org/2010/354.pdf
	c0 := api.Add(x.A0, x.A1)
	c2 := api.Mul(x.A1, ext.uSquare)
	c2 = api.Add(c2, x.A0)

	c0 = api.Mul(c0, c2) // (x1+x2)*(x1+(u**2)x2)
	c2 = api.Mul(x.A0, x.A1)
	c2 = api.Add(c2, c2)
	e.A1 = c2
	c2 = api.Add(c2, c2)
	e.A0 = api.Add(c0, c2)

	return e
}

// DivByFp divides an fp2 elmt by an fp elmt
func (e *E2) DivByFp(api frontend.API, e1 E2, c interface{}) *E2 {
	e.A0 = api.DivUnchecked(e1.A0, c)
	e.A1 = api.DivUnchecked(e1.A1, c)
	return e
}

// MulByFp multiplies an fp2 elmt by an fp elmt
func (e *E2) MulByFp(api frontend.API, e1 E2, c interface{}) *E2 {
	e.A0 = api.Mul(e1.A0, c)
	e.A1 = api.Mul(e1.A1, c)
	return e
}

// MulByNonResidue multiplies an fp2 elmt by the imaginary elmt
// ext.uSquare is the square of the imaginary root
func (e *E2) MulByNonResidue(api frontend.API, e1 E2) *E2 {
	x := e1.A0
	e.A0 = api.Mul(e1.A1, ext.uSquare)
	e.A1 = x
	return e
}

// Conjugate conjugation of an e2 elmt
func (e *E2) Conjugate(api frontend.API, e1 E2) *E2 {
	e.A0 = e1.A0
	e.A1 = api.Sub(0, e1.A1)
	return e
}

var InverseE2Hint = func(curve ecc.ID, inputs []*big.Int, res []*big.Int) error {
	var a, c bls12377.E2

	a.A0.SetBigInt(inputs[0])
	a.A1.SetBigInt(inputs[1])

	c.Inverse(&a)

	c.A0.ToBigIntRegular(res[0])
	c.A1.ToBigIntRegular(res[1])

	return nil
}

func init() {
	hint.Register(InverseE2Hint)
}

// Inverse e2 elmts
func (e *E2) Inverse(api frontend.API, e1 E2) *E2 {

	res, err := api.NewHint(InverseE2Hint, 2, e1.A0, e1.A1)
	if err != nil {
		// err is non-nil only for invalid number of inputs
		panic(err)
	}

	var e3, one E2
	e3.assign(res[:2])
	one.SetOne()

	// 1 == e3 * e1
	e3.Mul(api, e3, e1)
	e3.MustBeEqual(api, one)

	e.assign(res[:2])

	return e
}

var DivE2Hint = func(curve ecc.ID, inputs []*big.Int, res []*big.Int) error {
	var a, b, c bls12377.E2

	a.A0.SetBigInt(inputs[0])
	a.A1.SetBigInt(inputs[1])
	b.A0.SetBigInt(inputs[2])
	b.A1.SetBigInt(inputs[3])

	c.Inverse(&b).Mul(&c, &a)

	c.A0.ToBigIntRegular(res[0])
	c.A1.ToBigIntRegular(res[1])

	return nil
}

func init() {
	hint.Register(DivE2Hint)
}

// DivUnchecked e2 elmts
func (e *E2) DivUnchecked(api frontend.API, e1, e2 E2) *E2 {

	res, err := api.NewHint(DivE2Hint, 2, e1.A0, e1.A1, e2.A0, e2.A1)
	if err != nil {
		// err is non-nil only for invalid number of inputs
		panic(err)
	}

	var e3 E2
	e3.assign(res[:2])

	// e1 == e3 * e2
	e3.Mul(api, e3, e2)
	e3.MustBeEqual(api, e1)

	e.assign(res[:2])

	return e
}

// Assign a value to self (witness assignment)
func (e *E2) Assign(a *bls12377.E2) {
	e.A0 = (fr.Element)(a.A0)
	e.A1 = (fr.Element)(a.A1)
}

// MustBeEqual constraint self to be equal to other into the given constraint system
func (e *E2) MustBeEqual(api frontend.API, other E2) {
	api.AssertIsEqual(e.A0, other.A0)
	api.AssertIsEqual(e.A1, other.A1)
}
