package retriever

type Options struct {
	// Top K documents to retrieve
	K                  int //FIXME: maybe we don't want to statically set this
	ReciprocalWeight   float64
	ReciprocalPriority int
}

func (o *Options) Apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func (o *Options) Weight() float64 {
	return o.ReciprocalWeight
}

func (o *Options) Priority() int {
	return o.ReciprocalPriority
}

type Option func(*Options)

func WithK(k int) Option {
	return func(o *Options) {
		o.K = k
	}
}

func WithReciprocalWeight(weight float64) Option {
	return func(o *Options) {
		o.ReciprocalWeight = weight
	}
}

func WithReciprocalPriority(priority int) Option {
	return func(o *Options) {
		o.ReciprocalPriority = priority
	}
}
