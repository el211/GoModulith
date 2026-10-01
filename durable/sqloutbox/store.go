package sqloutbox

import (
 "context"
 "database/sql"
 "errors"
 "fmt"
 "strings"
 "time"
 "github.com/el211/GoModulith/durable"
 "github.com/el211/GoModulith/durable/reliable"
)
// Dialect specifies the supported SQL placeholder and conflict syntax.
type Dialect string
const (SQLite Dialect="sqlite";Postgres Dialect="postgres")
var ErrLease=errors.New("sqloutbox: lease not owned or message missing")
// Store is a SQL-backed reliable outbox.
type Store struct{DB *sql.DB;Dialect Dialect}
func(s Store)bind(i int)string{if s.Dialect==Postgres{return fmt.Sprintf("$%d",i)};return "?"}
func(s Store)check()error{if s.DB==nil{return errors.New("sqloutbox: nil DB")};if s.Dialect!=SQLite&&s.Dialect!=Postgres{return errors.New("sqloutbox: unsupported dialect")};return nil}
// Schema contains portable DDL. Timestamp columns store UTC Unix milliseconds.
// Install with DB.Exec(ctx, Schema(dialect)) or an application's migrations.
func Schema(d Dialect)string{
 id:="TEXT";payload:="BLOB";if d==Postgres{payload="BYTEA"}
 return "CREATE TABLE IF NOT EXISTS gomodulith_outbox ("+
 "id "+id+" PRIMARY KEY, topic TEXT NOT NULL, payload "+payload+" NOT NULL, created_at BIGINT NOT NULL, "+
 "attempts INTEGER NOT NULL DEFAULT 0, available_at BIGINT NOT NULL, lease_until BIGINT NOT NULL DEFAULT 0, "+
 "owner TEXT NOT NULL DEFAULT '', state TEXT NOT NULL DEFAULT 'pending', last_error TEXT NOT NULL DEFAULT ''"+
 "); CREATE INDEX IF NOT EXISTS gomodulith_outbox_claim_idx ON gomodulith_outbox(state,available_at,lease_until);"
}
// EnqueueTx inserts an event inside an existing business transaction.
func(s Store)EnqueueTx(ctx context.Context,tx *sql.Tx,e durable.Event)error{
 if err:=s.check();err!=nil{return err};if tx==nil||e.ID==""||e.Topic==""{return errors.New("sqloutbox: transaction, ID and topic required")}
 now:=e.CreatedAt;if now.IsZero(){now=time.Now().UTC()}
 q:="INSERT INTO gomodulith_outbox(id,topic,payload,created_at,available_at) VALUES ("+
 strings.Join([]string{s.bind(1),s.bind(2),s.bind(3),s.bind(4),s.bind(5)},",")+")"
 _,err:=tx.ExecContext(ctx,q,e.ID,e.Topic,e.Payload,now.UnixMilli(),now.UnixMilli());return err
}
// Enqueue opens a standalone transaction. For atomic business changes use EnqueueTx.
func(s Store)Enqueue(ctx context.Context,e durable.Event)error{
 if err:=s.check();err!=nil{return err}
 tx,err:=s.DB.BeginTx(ctx,nil);if err!=nil{return err};defer tx.Rollback()
 if err:=s.EnqueueTx(ctx,tx,e);err!=nil{return err};return tx.Commit()
}
// Claim atomically acquires a pending message using compare-and-swap updates.
// Postgres and SQLite both support this portable approach. Parallel workers
// may scan the same candidate but only one can claim the row.
func(s Store)Claim(ctx context.Context,worker string,now time.Time,lease time.Duration,limit int)([]reliable.Message,error){
 if err:=s.check();err!=nil{return nil,err};if worker==""||lease<=0||limit<=0{return nil,errors.New("sqloutbox: invalid claim arguments")}
 query:="SELECT id FROM gomodulith_outbox WHERE state='pending' AND available_at<="+s.bind(1)+" AND lease_until<="+s.bind(2)+" ORDER BY available_at,id LIMIT "+s.bind(3)
 rows,err:=s.DB.QueryContext(ctx,query,now.UnixMilli(),now.UnixMilli(),limit);if err!=nil{return nil,err}
 ids:=[]string{};for rows.Next(){var id string;if err:=rows.Scan(&id);err!=nil{rows.Close();return nil,err};ids=append(ids,id)}
 if err:=rows.Err();err!=nil{rows.Close();return nil,err};rows.Close()
 result:=make([]reliable.Message,0,len(ids))
 for _,id:=range ids{
  q:="UPDATE gomodulith_outbox SET owner="+s.bind(1)+",lease_until="+s.bind(2)+" WHERE id="+s.bind(3)+" AND state='pending' AND available_at<="+s.bind(4)+" AND lease_until<="+s.bind(5)
  changed,err:=s.DB.ExecContext(ctx,q,worker,now.Add(lease).UnixMilli(),id,now.UnixMilli(),now.UnixMilli());if err!=nil{return result,err}
  n,err:=changed.RowsAffected();if err!=nil{return result,err};if n!=1{continue}
  var e reliable.Message;var created,available,until int64
  q="SELECT id,topic,payload,created_at,attempts,available_at,lease_until,last_error FROM gomodulith_outbox WHERE id="+s.bind(1)+" AND owner="+s.bind(2)
  if err:=s.DB.QueryRowContext(ctx,q,id,worker).Scan(&e.ID,&e.Topic,&e.Payload,&created,&e.Attempts,&available,&until,&e.LastError);err!=nil{return result,err}
  e.CreatedAt=time.UnixMilli(created).UTC();e.AvailableAt=time.UnixMilli(available).UTC();e.LeaseUntil=time.UnixMilli(until).UTC();result=append(result,e)
 }
 return result,nil
}
func(s Store)change(ctx context.Context,q string,args ...any)error{
 if err:=s.check();err!=nil{return err};res,err:=s.DB.ExecContext(ctx,q,args...);if err!=nil{return err};n,err:=res.RowsAffected();if err!=nil{return err};if n!=1{return ErrLease};return nil
}
func(s Store)Ack(ctx context.Context,id,worker string)error{
 q:="DELETE FROM gomodulith_outbox WHERE id="+s.bind(1)+" AND owner="+s.bind(2)+" AND state='pending'"
 return s.change(ctx,q,id,worker)
}
func(s Store)Retry(ctx context.Context,id,worker string,attempt int,available time.Time,reason string)error{
 q:="UPDATE gomodulith_outbox SET attempts="+s.bind(1)+", available_at="+s.bind(2)+", last_error="+s.bind(3)+", owner='',lease_until=0 WHERE id="+s.bind(4)+" AND owner="+s.bind(5)+" AND state='pending'"
 return s.change(ctx,q,attempt,available.UnixMilli(),reason,id,worker)
}
func(s Store)Dead(ctx context.Context,id,worker string,attempt int,reason string)error{
 q:="UPDATE gomodulith_outbox SET state='dead',attempts="+s.bind(1)+",last_error="+s.bind(2)+",owner='',lease_until=0 WHERE id="+s.bind(3)+" AND owner="+s.bind(4)+" AND state='pending'"
 return s.change(ctx,q,attempt,reason,id,worker)
}
var _ reliable.Store=Store{}
