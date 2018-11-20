package httpx

// Middleware for a piece of middleware.
// Some middleware use this middleware out of the box,
// so in most cases you can just pass somepackage.New
type Middleware func(Handler) Handler

// Chain acts as a list of Handler middlewares.
// Chain is effectively immutable:
// once created, it will always hold
// the same set of middlewares in the same order.
type Chain struct {
	middlewares []Middleware
}

// NewChain creates a new chain,
// memorizing the given list of middleware middlewares.
// New serves no other function,
// middlewares are only called upon a call to Then().
func NewChain(middlewares ...Middleware) Chain {
	if len(middlewares) == 0 {
		return Chain{middlewares: []Middleware{}}
	}
	return Chain{middlewares: middlewares}
}

// Then chains the middleware and returns the final Handler.
//     New(m1, m2, m3).Then(h)
// is equivalent to:
//     m1(m2(m3(h)))
// When the request comes in, it will be passed to m1, then m2, then m3
// and finally, the given handler
// (assuming every middleware calls the following one).
//
// A chain can be safely reused by calling Then() several times.
//     stdStack := alice.New(ratelimitHandler, csrfHandler)
//     indexPipe = stdStack.Then(indexHandler)
//     authPipe = stdStack.Then(authHandler)
// Note that middlewares are called on every call to Then()
// and thus several instances of the same middleware will be created
// when a chain is reused in this way.
// For proper middleware, this should cause no problems.
//
// Then() treats nil as http.DefaultServeMux.
func (c Chain) Then(h Handler) Handler {
	if len(c.middlewares) == 0 {
		return h
	}

	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}

	return h
}

// ThenFunc works identically to Then, but takes
// a HandlerFunc instead of a Handler.
//
// The following two statements are equivalent:
//     c.Then(HandlerFunc(fn))
//     c.ThenFunc(fn)
//
// ThenFunc provides all the guarantees of Then.
func (c Chain) ThenFunc(fn HandlerFunc) Handler {
	return c.Then(fn)
}

// Append extends a chain, adding the specified middlewares
// as the last ones in the request flow.
//
// Append returns a new chain, leaving the original one untouched.
//
//     stdChain := alice.New(m1, m2)
//     extChain := stdChain.Append(m3, m4)
//     // requests in stdChain go m1 -> m2
//     // requests in extChain go m1 -> m2 -> m3 -> m4
func (c Chain) Append(middlewares ...Middleware) Chain {
	newCons := make([]Middleware, 0, len(c.middlewares)+len(middlewares))
	newCons = append(newCons, c.middlewares...)
	newCons = append(newCons, middlewares...)
	return Chain{newCons}
}

// Extend extends a chain by adding the specified chain
// as the last one in the request flow.
//
// Extend returns a new chain, leaving the original one untouched.
//
//     stdChain := alice.New(m1, m2)
//     ext1Chain := alice.New(m3, m4)
//     ext2Chain := stdChain.Extend(ext1Chain)
//     // requests in stdChain go  m1 -> m2
//     // requests in ext1Chain go m3 -> m4
//     // requests in ext2Chain go m1 -> m2 -> m3 -> m4
//
// Another example:
//  aHtmlAfterNosurf := alice.New(m2)
// 	aHtml := alice.New(m1, func(h Handler) Handler {
// 		csrf := nosurf.New(h)
// 		csrf.SetFailureHandler(aHtmlAfterNosurf.ThenFunc(csrfFail))
// 		return csrf
// 	}).Extend(aHtmlAfterNosurf)
//		// requests to aHtml hitting nosurfs success handler go m1 -> nosurf -> m2 -> target-handler
//		// requests to aHtml hitting nosurfs failure handler go m1 -> nosurf -> m2 -> csrfFail
func (c Chain) Extend(chain Chain) Chain {
	return c.Append(chain.middlewares...)
}
