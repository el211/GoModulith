package fiber
import(
 "github.com/el211/GoModulith"
 fiberlib "github.com/gofiber/fiber/v2"
)
const key="gomodulith.application"
// Middleware attaches app to Fiber-local request storage.
func Middleware(app *modulith.App)fiberlib.Handler{return func(c *fiberlib.Ctx)error{c.Locals(key,app);return c.Next()}}
// FromContext reads the application attached by Middleware.
func FromContext(c *fiberlib.Ctx)(*modulith.App,bool){v,ok:=c.Locals(key).(*modulith.App);return v,ok}
