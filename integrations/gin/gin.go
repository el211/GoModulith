package gin
import(
 "context"
 "github.com/el211/GoModulith"
 ginlib "github.com/gin-gonic/gin"
)
type contextKey struct{}
// Middleware exposes the application on request.Context.
func Middleware(app *modulith.App)ginlib.HandlerFunc{
 return func(c *ginlib.Context){c.Request=c.Request.WithContext(context.WithValue(c.Request.Context(),contextKey{},app));c.Next()}
}
// FromContext retrieves the registered application.
func FromContext(ctx context.Context)(*modulith.App,bool){v,ok:=ctx.Value(contextKey{}).(*modulith.App);return v,ok}
