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

func TestOptions_AddFooterFile(t *testing.T) {
	its := assert.New(t)

	t.Run("do nothing when path is empty", func(t *testing.T) {
		opts := PdfOptions().AddFooterFile("")

		its.Nil(opts.footer)
	})

	t.Run("add css file path", func(t *testing.T) {
		opts := PdfOptions().AddFooterFile("footer")

		its.Equal("footer", *opts.footer)
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
		opt3 := PdfOptions().AddFooterFile("footer").SetLandscape(true).SetPaperSize(100, 50)

		result := MergeOptions(opt, opt2, opt3)
		expect := PdfOptions().AddImageFile("imageFile").AddCssFile("cssFile").SetMargin(10, 10, 50, 10).AddFooterFile("footer").SetLandscape(true).SetPaperSize(100, 50)

		its.Equal(expect, result)
	})
}
