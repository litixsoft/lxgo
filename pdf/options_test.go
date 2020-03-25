package lxPdf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_AddCssFile(t *testing.T) {
	its := assert.New(t)

	t.Run("do nothing when path is empty", func(t *testing.T) {
		opts := PdfOptions().AddCssFile("")

		its.Nil(opts.css)
	})

	t.Run("add css file path", func(t *testing.T) {
		opts := PdfOptions().AddCssFile("testFilePath")

		its.Equal([]string{"testFilePath"}, *opts.css)
	})

	t.Run("add two css file path", func(t *testing.T) {
		opts := PdfOptions().AddCssFile("testFilePath")

		// add second css file
		opts.AddCssFile("testFilePath2")

		its.Equal([]string{"testFilePath", "testFilePath2"}, *opts.css)
	})
}

func TestOptions_AddHeader(t *testing.T) {
	its := assert.New(t)

	t.Run("do nothing when path is empty", func(t *testing.T) {
		opts := PdfOptions().AddHeader("")

		its.Nil(opts.header)
		its.Nil(opts.headerData)
	})

	t.Run("add header file path", func(t *testing.T) {
		opts := PdfOptions().AddHeader("header")

		its.Equal("header", *opts.header)
		its.Nil(opts.headerData)
	})

	t.Run("do nothing when header path is empty although header data is not empty", func(t *testing.T) {
		data := map[string]interface{}{
			"foo": "bar",
		}
		opts := PdfOptions().AddHeader("", &data)

		its.Nil(opts.header)
		its.Nil(opts.headerData)
	})

	t.Run("add header file path and header data", func(t *testing.T) {
		data := map[string]interface{}{
			"foo": "bar",
		}
		opts := PdfOptions().AddHeader("header", &data)

		its.Equal("header", *opts.header)
		its.Equal(data, *opts.headerData)
	})
}

func TestOptions_AddFooter(t *testing.T) {
	its := assert.New(t)

	t.Run("do nothing when path is empty", func(t *testing.T) {
		opts := PdfOptions().AddFooter("")

		its.Nil(opts.footer)
		its.Nil(opts.footerData)
	})

	t.Run("add footer file path", func(t *testing.T) {
		opts := PdfOptions().AddFooter("footer")

		its.Equal("footer", *opts.footer)
		its.Nil(opts.footerData)
	})

	t.Run("do nothing when footer path is empty although footer data is not empty", func(t *testing.T) {
		data := map[string]interface{}{
			"foo": "bar",
		}
		opts := PdfOptions().AddFooter("", &data)

		its.Nil(opts.footer)
		its.Nil(opts.footerData)
	})

	t.Run("add footer file path and footer data", func(t *testing.T) {
		data := map[string]interface{}{
			"foo": "bar",
		}
		opts := PdfOptions().AddFooter("footer", &data)

		its.Equal("footer", *opts.footer)
		its.Equal(data, *opts.footerData)
	})
}

func TestOptions_AddImageFile(t *testing.T) {
	its := assert.New(t)

	t.Run("do nothing when path is empty", func(t *testing.T) {
		opts := PdfOptions().AddImageFile("")

		its.Nil(opts.images)
	})

	t.Run("add image file path", func(t *testing.T) {
		opts := PdfOptions().AddImageFile("imageFile")

		its.Equal([]string{"imageFile"}, *opts.images)
	})

	t.Run("add two css file path", func(t *testing.T) {
		opts := PdfOptions().AddImageFile("imageFile")

		// add second css file
		opts.AddImageFile("imageFile2")

		its.Equal([]string{"imageFile", "imageFile2"}, *opts.images)
	})
}

func TestOptions_SetLandscape(t *testing.T) {
	its := assert.New(t)

	t.Run("do nothing when flag is false", func(t *testing.T) {
		opts := PdfOptions().SetLandscape(false)

		its.Nil(opts.opts)
	})

	t.Run("set orientation to landscape", func(t *testing.T) {
		opts := PdfOptions().SetLandscape(true)

		its.NotNil(opts.opts)
		its.Equal(true, opts.opts.Landscape)
	})
}

func TestOptions_SetMargin(t *testing.T) {
	its := assert.New(t)

	t.Run("do nothing when values are lower or equal zero", func(t *testing.T) {
		opts := PdfOptions().SetMargin(0, -1, 0, -1)

		its.Nil(opts.opts)
	})

	t.Run("set only bottom margin", func(t *testing.T) {
		opts := PdfOptions().SetMargin(0, 0, 10, 0)

		its.NotNil(opts.opts)
		its.IsType(margin{}, *opts.opts.Margin)
		its.Equal(float64(10), opts.opts.Margin.Bottom)
	})

	t.Run("set all margin values", func(t *testing.T) {
		opts := PdfOptions().SetMargin(10, 20, 30, 40)

		expect := margin{
			Bottom: 30,
			Left:   40,
			Right:  20,
			Top:    10,
		}
		its.NotNil(opts.opts)
		its.IsType(expect, *opts.opts.Margin)
		its.Equal(expect, *opts.opts.Margin)
	})
}

func TestOptions_SetPaperSize(t *testing.T) {
	its := assert.New(t)

	t.Run("do nothing when values are lower or equal zero", func(t *testing.T) {
		opts := PdfOptions().SetPaperSize(0, -1)

		its.Nil(opts.opts)
	})

	t.Run("set only page height value", func(t *testing.T) {
		opts := PdfOptions().SetPaperSize(20, 0)

		expect := paperSize{
			Height: 20,
		}

		its.NotNil(opts.opts)
		its.IsType(expect, *opts.opts.PaperSize)
		its.Equal(expect, *opts.opts.PaperSize)
	})

	t.Run("set all paper size values", func(t *testing.T) {
		opts := PdfOptions().SetPaperSize(123, 45)

		expect := paperSize{
			Height: 123,
			Width:  45,
		}

		its.NotNil(opts.opts)
		its.IsType(expect, *opts.opts.PaperSize)
		its.Equal(expect, *opts.opts.PaperSize)
	})
}

func TestMergeOptions(t *testing.T) {
	its := assert.New(t)

	t.Run("set a few fields", func(t *testing.T) {
		opt := PdfOptions().AddImageFile("imageFile").AddCssFile("cssFile")
		opt2 := PdfOptions().SetMargin(0, 10, 50, 0)

		result := MergeOptions(opt, opt2)
		expect := PdfOptions().AddImageFile("imageFile").AddCssFile("cssFile").SetMargin(0, 10, 50, 0)

		its.Equal(expect, result)
	})

	t.Run("set all fields", func(t *testing.T) {
		opt := PdfOptions().AddImageFile("imageFile").AddCssFile("cssFile")
		opt2 := PdfOptions().SetMargin(10, 10, 50, 10)
		opt3 := PdfOptions().AddFooter("footer").SetLandscape(true).SetPaperSize(100, 50)

		result := MergeOptions(opt, opt2, opt3)
		expect := PdfOptions().AddImageFile("imageFile").AddCssFile("cssFile").SetMargin(10, 10, 50, 10).AddFooter("footer").SetLandscape(true).SetPaperSize(100, 50)

		its.Equal(expect, result)
	})
}
