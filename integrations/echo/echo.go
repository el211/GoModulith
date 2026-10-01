package echo
import(
 "context"
 "github.com/el211/GoModulith"
 echolib "github.com/labstack/echo/v4"
)
type contextKey struct{}
// Middleware attaches app to the request's context.
func Middleware(app *modulith.App)echolib.MiddlewareFunc{return func(next echolib.HandlerFunc)echolib.HandlerFunc{
 return func(c echolib.Context)error{
  c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(),contextKey{},app)))
  return next(c)
 }
}}
func FromContext(ctx context.Context)(*modulith.App,bool){v,ok:=ctx.Value(contextKey{}).(*modulith.App);return v,ok}
