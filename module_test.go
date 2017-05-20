package lsx_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/akutz/lsx"
)

const (
	// maxModuleType is the maximum module type constant.
	maxModuleType = lsx.VolumeModuleType
)

func TestModule(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Module Suite")
}

var _ = Describe("Module", func() {

	var (
		err      error
		toParse  interface{}
		actual   lsx.ModuleType
		expected lsx.ModuleType
	)

	JustBeforeEach(func() {
		actual, err = lsx.ParseModuleType(toParse)
	})
	AfterEach(func() {
		err = nil
		toParse = nil
		actual = 0
		expected = 0
	})

	assertValid := func(v interface{}, x lsx.ModuleType) {
		Context(fmt.Sprintf("%[1]T(%[1]v)", v), func() {
			BeforeEach(func() {
				toParse = v
				expected = x
			})
			It("valid module type", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(actual).Should(Equal(expected))
			})
		})
		pv := addrOf(v)
		if pv == nil {
			return
		}
		Context(fmt.Sprintf("%[1]T(%[2]v)", pv, v), func() {
			BeforeEach(func() {
				toParse = pv
				expected = x
			})
			It("valid module type", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Ω(actual).Should(Equal(expected))
			})
		})
	}

	assertInvalid := func(v interface{}) {
		Context(fmt.Sprintf("%[1]T(%[1]v)", v), func() {
			BeforeEach(func() {
				toParse = v
				expected = 0
			})
			It("invalid module type", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err.Error()).Should(Equal(
					fmt.Sprintf("error: invalid module type: %v", v)))
				Ω(actual).Should(Equal(expected))
			})
		})
		pv := addrOf(v)
		if pv == nil {
			return
		}
		Context(fmt.Sprintf("%[1]T(%[2]v)", pv, v), func() {
			BeforeEach(func() {
				toParse = pv
				expected = 0
			})
			It("invalid module type", func() {
				Ω(err).Should(HaveOccurred())
				Ω(err.Error()).Should(Equal(
					fmt.Sprintf("error: invalid module type: %v", v)))
				Ω(actual).Should(Equal(expected))
			})
		})
	}

	// handle valid parse scenarios
	for mt := lsx.InvalidModuleType + 1; mt <= maxModuleType; mt++ {
		szMT := mt.String()
		szMTUC := strings.ToUpper(szMT)
		assertValid(szMT, mt)
		assertValid(szMTUC, mt)

		assertValid(&modTypeStringer{szMT}, mt)
		assertValid(&modTypeStringer{szMTUC}, mt)

		assertValid(uint(mt), mt)
		assertValid(uint8(mt), mt)
		assertValid(uint16(mt), mt)
		assertValid(uint32(mt), mt)
		assertValid(uint64(mt), mt)

		assertValid(int(mt), mt)
		assertValid(int8(mt), mt)
		assertValid(int16(mt), mt)
		assertValid(int32(mt), mt)
		assertValid(int64(mt), mt)

		assertValid(float32(mt), mt)
		assertValid(float64(mt), mt)

		// TODO complex64
		// TODO complex128
	}

	// handle invalid parse scenarios
	assertInvalid("volumeService")
	assertInvalid(int8(-1))
	assertInvalid(-2)
	assertInvalid(1.5)
	assertInvalid("")
	assertInvalid(nil)
})

type modTypeStringer struct {
	s string
}

func (m modTypeStringer) String() string {
	return m.s
}

func addrOf(v interface{}) interface{} {
	switch tv := v.(type) {
	case string:
		return &tv
	case uint:
		return &tv
	case uint8:
		return &tv
	case uint16:
		return &tv
	case uint32:
		return &tv
	case uint64:
		return &tv
	case int:
		return &tv
	case int8:
		return &tv
	case int16:
		return &tv
	case int32:
		return &tv
	case int64:
		return &tv
	case float32:
		return &tv
	case float64:
		return &tv
	}
	return nil
}
