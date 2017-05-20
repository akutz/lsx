package lsx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/akutz/lsx"
)

func init() {
	exampleConfigJSON, _ = ioutil.ReadFile("config_example.json")
}

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config", func() {

	var (
		ctx    context.Context
		config lsx.Config
	)

	BeforeEach(func() {
		ctx = context.Background()
		config = lsx.Config{}
		Ω(json.Unmarshal(
			exampleConfigJSON,
			&config)).ShouldNot(HaveOccurred())
	})
	AfterEach(func() {
		config = nil
	})

	It("should have len(4)", func() {
		Ω(config.Len()).Should(Equal(4))
	})
	It("should have debug log level", func() {
		lvl := config.Get(ctx, "logging.level")
		Ω(lvl).Should(BeAssignableToTypeOf(typeOfString))
		Ω(lvl).Should(Equal("debug"))
	})
	It("should marshal to minified JSON", func() {
		buf, err := json.Marshal(config)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf).Should(MatchJSON(exampleConfigJSON))
		Ω(buf).ShouldNot(HavePrefix("{\n"))
		Ω(buf).ShouldNot(HaveSuffix("\n}"))
	})
	It("should marshal to prettified JSON", func() {
		buf, err := json.MarshalIndent(config, "", "  ")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf).Should(MatchJSON(exampleConfigJSON))
		Ω(buf).Should(HavePrefix("{\n"))
		Ω(buf).Should(HaveSuffix("\n}"))
	})
	It("should marshal to indentable JSON", func() {
		buf, err := json.Marshal(config)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf).Should(MatchJSON(exampleConfigJSON))
		Ω(buf).ShouldNot(HavePrefix("{\n"))
		Ω(buf).ShouldNot(HaveSuffix("\n}"))
		dst := &bytes.Buffer{}
		err = json.Indent(dst, buf, "", "  ")
		Ω(err).ShouldNot(HaveOccurred())
		buf = dst.Bytes()
		Ω(buf).Should(MatchJSON(exampleConfigJSON))
		Ω(buf).Should(HavePrefix("{\n"))
		Ω(buf).Should(HaveSuffix("\n}"))
	})
	It("should marshal to compactable JSON", func() {
		buf, err := json.MarshalIndent(config, "", "  ")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(buf).Should(MatchJSON(exampleConfigJSON))
		Ω(buf).Should(HavePrefix("{\n"))
		Ω(buf).Should(HaveSuffix("\n}"))
		dst := &bytes.Buffer{}
		err = json.Compact(dst, buf)
		Ω(err).ShouldNot(HaveOccurred())
		buf = dst.Bytes()
		Ω(buf).Should(MatchJSON(exampleConfigJSON))
		Ω(buf).ShouldNot(HavePrefix("{\n"))
		Ω(buf).ShouldNot(HaveSuffix("\n}"))
	})

	Context("scoped to service svc00", func() {
		BeforeEach(func() {
			config = config.Scope(ctx, "services.svc00")
		})
		It("should have len(4)", func() {
			Ω(config.Len()).Should(Equal(4))
		})
		It("should be the json for just the service", func() {
			buf, err := json.Marshal(config)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(buf).Should(MatchJSON(svc00ConfigJSON))
		})
		It("should have info log level", func() {
			lvl := config.Get(ctx, "logging.level")
			Ω(lvl).Should(BeAssignableToTypeOf(typeOfString))
			Ω(lvl).Should(Equal("info"))
		})

		Context("with LSX_LOGGING_LEVEL set", func() {
			BeforeEach(func() {
				os.Setenv("LSX_LOGGING_LEVEL", "debug")
			})
			AfterEach(func() {
				os.Setenv("LSX_LOGGING_LEVEL", "")
			})
			It("should have debug log level", func() {
				lvl := config.Get(ctx, "logging.level")
				Ω(lvl).Should(BeAssignableToTypeOf(typeOfString))
				Ω(lvl).Should(Equal("debug"))
			})
		})
	})

	Context("with a struct member", func() {
		BeforeEach(func() {
			config["data"] = &testStruct{
				Name: "hello",
				World: map[string]interface{}{
					"logLevel": 10,
				},
			}
		})
		Specify("data.name", func() {
			v := config.Get(ctx, "data.name")
			Ω(v).Should(BeAssignableToTypeOf(""))
			Ω(v).Should(Equal("hello"))
		})
		Specify("data.world.logLevel", func() {
			v := config.Get(ctx, "data.world.logLevel")
			Ω(v).Should(BeAssignableToTypeOf(10))
			Ω(v).Should(Equal(10))
		})
	})

	Context("with an array", func() {
		var array []interface{}
		var inmap map[string]interface{}

		BeforeEach(func() {
			inmap = map[string]interface{}{
				"name": "c3p0",
			}
			array = []interface{}{
				1,
				"two",
				inmap,
			}
			config["array"] = array
		})
		AfterEach(func() {
			inmap = nil
			array = nil
		})

		Specify("array.c3p0 is a map", func() {
			v := config.Get(ctx, "array.c3p0")
			Ω(v).Should(BeAssignableToTypeOf(map[string]interface{}{}))
			m, _ := v.(map[string]interface{})
			Ω(m).Should(HaveLen(1))
			Ω(m["name"]).Should(Equal("c3p0"))
		})

		Context("with a struct member", func() {
			BeforeEach(func() {
				array = append(array, &testStruct{
					Name: "hello",
					World: map[string]interface{}{
						"logLevel": 10,
					},
				})
				config["array"] = array
				inmap["myTestStruct"] = &testStruct{
					Name: "hello",
					World: map[string]interface{}{
						"logLevel": 5,
					},
				}
			})
			Specify("array.hello.world.loglevel", func() {
				v := config.Get(ctx, "array.hello.world.loglevel")
				Ω(v).Should(BeAssignableToTypeOf(10))
				Ω(v).Should(Equal(10))
			})
			Specify("array.c3p0.myTestStruct.world.loglevel", func() {
				v := config.Get(ctx, "array.c3p0.myTestStruct.world.loglevel")
				Ω(v).Should(BeAssignableToTypeOf(5))
				Ω(v).Should(Equal(5))
			})
		})
	})

	Context("scoped to service svc01", func() {
		BeforeEach(func() {
			config = config.Scope(ctx, "services.svc01")
		})
		It("should fail", func() {
			Ω(config).Should(BeEmpty())
		})
	})

	Context("scoped to server svr00", func() {
		BeforeEach(func() {
			config = config.Scope(ctx, "servers.svr00")
		})
		It("should have len(3)", func() {
			Ω(config.Len()).Should(Equal(3))
		})
		It("should have debug log level", func() {
			lvl := config.Get(ctx, "logging.level")
			Ω(lvl).Should(BeAssignableToTypeOf(typeOfString))
			Ω(lvl).Should(Equal("debug"))
		})
		Context("with addrs", func() {
			var (
				addrs   interface{}
				szAddrs string
			)
			BeforeEach(func() {
				addrs = config.Get(ctx, "addrs")
				szAddrs = config.GetStr(ctx, "addrs")
			})
			AfterEach(func() {
				addrs = nil
				szAddrs = ""
			})
			It("should have len(1)", func() {
				Ω(addrs).Should(BeAssignableToTypeOf(typeOfArrInterface))
				Ω(addrs).Should(HaveLen(1))
			})
			It("should marshal to JSON when retrieved via GetStr", func() {
				Ω(szAddrs).Should(MatchJSON(svr00Addrs))
			})
		})
	})

	Context("scoped to server svr01", func() {
		BeforeEach(func() {
			config = config.Scope(ctx, "servers.svr01")
		})
		It("should have len(3)", func() {
			Ω(config.Len()).Should(Equal(3))
		})
		It("should have debug log level", func() {
			lvl := config.Get(ctx, "logging.level")
			Ω(lvl).Should(BeAssignableToTypeOf(typeOfString))
			Ω(lvl).Should(Equal("debug"))
		})
		Context("with addrs", func() {
			var (
				addrs   interface{}
				szAddrs string
			)
			BeforeEach(func() {
				addrs = config.Get(ctx, "addrs")
				szAddrs = config.GetStr(ctx, "addrs")
			})
			AfterEach(func() {
				addrs = nil
				szAddrs = ""
			})
			It("should have len(2)", func() {
				Ω(addrs).Should(BeAssignableToTypeOf(typeOfArrInterface))
				Ω(addrs).Should(HaveLen(2))
			})
			It("should marshal to JSON when retrieved via GetStr", func() {
				Ω(szAddrs).Should(MatchJSON(svr01Addrs))
			})
		})
	})
})

const (
	typeOfString = ``

	svr00Addrs = `["tcp://127.0.0.1:7979"]`
	svr01Addrs = `["tcp://127.0.0.1:8989","unix:///tmp/lsx/run/csi.sock"]`

	svc00ConfigJSON = `{"name":"svc00","servers":["svr00","svr01"],` +
		`"logging":{"level":"info","requests":true,"r` +
		`esponses":true},"api":{"volume":{"attach":{"` +
		`type":"vfs"},"mount":{"type":"libstorage","h` +
		`ost":"tcp://192.168.0.192:7979"}}}}`
	svc01ConfigJSON = `{}`
)

var (
	exampleConfigJSON  []byte
	typeOfArrInterface = []interface{}{}
)

type testStruct struct {
	Name  string
	World map[string]interface{}
}
