package render

type AsyncRender struct {
	frames chan *Frame
	render Render
}

func NewAsyncRender(render Render) *AsyncRender {
	return &AsyncRender{
		frames: make(chan *Frame, 1<<8),
		render: render,
	}
}

func (a *AsyncRender) Start() {
	for frame := range a.frames {
		a.render.Push(frame)
	}
}

func (a *AsyncRender) Push(frame *Frame) {
	a.frames <- frame
}
