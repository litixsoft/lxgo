package lxPdf

type options struct {
	css    *[]string
	footer *string
	images *[]string
	opts   *pdfOpts
}

type pdfOpts struct {
	Landscape bool       `json:"landscape,omitempty"`
	Margin    *margin    `json:"margin,omitempty"`
	PaperSize *paperSize `json:"paperSize,omitempty"`
}

type margin struct {
	Bottom float64 `json:"bottom,omitempty"`
	Left   float64 `json:"left,omitempty"`
	Right  float64 `json:"right,omitempty"`
	Top    float64 `json:"top,omitempty"`
}

type paperSize struct {
	Height float64 `json:"height,omitempty"`
	Width  float64 `json:"width,omitempty"`
}

func PdfOptions() *options {
	return &options{}
}

func (o *options) AddCssFile(path string) *options {
	if path != "" {
		if o.css == nil {
			o.css = new([]string)
		}
		*o.css = append(*o.css, path)
	}
	return o
}

func (o *options) AddFooterFile(path string) *options {
	if path != "" {
		o.footer = &path
	}
	return o
}

func (o *options) AddImageFile(path string) *options {
	if path != "" {
		if o.images == nil {
			o.images = new([]string)
		}
		*o.images = append(*o.images, path)
	}
	return o
}

func (o *options) SetLandscape(flag bool) *options {
	if flag {
		if o.opts == nil {
			o.opts = &pdfOpts{}
		}
		o.opts.Landscape = flag
	}
	return o
}

func (o *options) SetMargin(top, right, bottom, left float64) *options {
	if top > 0 || right > 0 || bottom > 0 || left > 0 {
		if o.opts == nil {
			o.opts = &pdfOpts{}
		}
		o.opts.Margin = &margin{
			Bottom: bottom,
			Left:   left,
			Right:  right,
			Top:    top,
		}
	}
	return o
}

func (o *options) SetPaperSize(height, width float64) *options {
	if height > 0 || width > 0 {
		if o.opts == nil {
			o.opts = &pdfOpts{}
		}
		o.opts.PaperSize = &paperSize{
			Height: height,
			Width:  width,
		}
	}
	return o
}

// MergeOptions combines the argued options into a single options in a last-one-wins fashion
func MergeOptions(opts ...*options) *options {
	mergedOpts := PdfOptions()

	for _, opt := range opts {
		// css
		if opt.css != nil {
			mergedOpts.css = opt.css
		}

		// images
		if opt.images != nil {
			mergedOpts.images = opt.images
		}

		// footer
		if opt.footer != nil {
			mergedOpts.footer = opt.footer
		}

		// opts
		if opt.opts != nil {
			if mergedOpts.opts == nil {
				mergedOpts.opts = &pdfOpts{}
			}

			// paperSize
			if opt.opts.PaperSize != nil {
				if mergedOpts.opts.PaperSize == nil {
					mergedOpts.opts.PaperSize = &paperSize{}
				}

				// height
				if opt.opts.PaperSize.Height != 0 {
					mergedOpts.opts.PaperSize.Height = opt.opts.PaperSize.Height
				}

				// width
				if opt.opts.PaperSize.Width != 0 {
					mergedOpts.opts.PaperSize.Width = opt.opts.PaperSize.Width
				}
			}

			// margin
			if opt.opts.Margin != nil {
				if mergedOpts.opts.Margin == nil {
					mergedOpts.opts.Margin = &margin{}
				}

				// marginTop
				if opt.opts.Margin.Top != 0 {
					mergedOpts.opts.Margin.Top = opt.opts.Margin.Top
				}

				// marginBottom
				if opt.opts.Margin.Bottom != 0 {
					mergedOpts.opts.Margin.Bottom = opt.opts.Margin.Bottom
				}

				// marginRight
				if opt.opts.Margin.Right != 0 {
					mergedOpts.opts.Margin.Right = opt.opts.Margin.Right
				}

				// marginLeft
				if opt.opts.Margin.Left != 0 {
					mergedOpts.opts.Margin.Left = opt.opts.Margin.Left
				}
			}

			// landscape
			if opt.opts.Landscape {
				mergedOpts.opts.Landscape = opt.opts.Landscape
			}
		}
	}
	return mergedOpts
}
