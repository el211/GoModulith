package grpc
import(
 "context"
 "github.com/el211/GoModulith"
 g "google.golang.org/grpc"
)
type contextKey struct{}
func FromContext(ctx context.Context)(*modulith.App,bool){v,ok:=ctx.Value(contextKey{}).(*modulith.App);return v,ok}
// UnaryServerInterceptor attaches app to unary RPC contexts.
func UnaryServerInterceptor(app *modulith.App)g.UnaryServerInterceptor{
 return func(ctx context.Context,req any,info *g.UnaryServerInfo,handler g.UnaryHandler)(any,error){
  return handler(context.WithValue(ctx,contextKey{},app),req)
 }
}
type wrappedStream struct{g.ServerStream;ctx context.Context}
func(w wrappedStream)Context()context.Context{return w.ctx}
// StreamServerInterceptor attaches app to streaming RPC contexts.
func StreamServerInterceptor(app *modulith.App)g.StreamServerInterceptor{
 return func(srv any,ss g.ServerStream,info *g.StreamServerInfo,handler g.StreamHandler)error{
  return handler(srv,wrappedStream{ServerStream:ss,ctx:context.WithValue(ss.Context(),contextKey{},app)})
 }
}
