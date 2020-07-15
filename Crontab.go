package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"gopkg.in/gomail.v2"
	"os"
	"strconv"
	"sync"
)

// https://www.cnblogs.com/jssyjam/p/11910851.html

//定时任务管理器
type Crontab struct {
	inner *cron.Cron
	ids   map[string]cron.EntryID
	mutex sync.Mutex
}

// new新的定时任务引擎
func NewCrontab() *Crontab {
	// cron.New() 默认从分开始，cron.WithSeconds() 加上后默认从秒开始
	return &Crontab{inner: cron.New(cron.WithSeconds()), ids: map[string]cron.EntryID{}}
}

// 引擎开始
func (c *Crontab) Start() {
	c.inner.Start()
}

// 停止引擎
func (c *Crontab) Stop() {
	c.inner.Stop()
}

/**
 * 删除任务
 * @param id 唯一任务id
 */
func (c *Crontab) DelByID(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	eid, ok := c.ids[id]
	if !ok {
		return
	}
	c.inner.Remove(eid)
}

/**
 * 实现接口的方式添加定时任务
 * @param id 唯一任务id
 * @param spec 配置定时执行时间
 * @param cj 需要执行的任务方法
 * @return error
 */
func (c *Crontab) AddJobByInterface(id string, spec string, cj cron.Job) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.ids[id]; ok {
		return errors.Errorf("crontab id exists")
	}
	eid, err := c.inner.AddJob(spec, cj)
	if err != nil {
		return err
	}
	c.ids[id] = eid
	return nil
}

/**
 * 添加函数作为定时任务
 * @param id 唯一任务id
 * @param spec 配置定时执行时间
 * @param f 需要执行的方法
 * @return error
 */
func (c *Crontab) AddJobByFunc(id string, spec string, f func()) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.ids[id]; ok {
		return errors.Errorf("crontab id exists")
	}
	eid, err := c.inner.AddFunc(spec, f)
	if err != nil {
		return err
	}
	c.ids[id] = eid
	return nil
}

/**
 * 判断是否存在任务
 * @param id 唯一任务id
 * @return bool
 */
func (c *Crontab) IsExistsJob(id string) bool {
	_, exist := c.ids[id]
	return exist
}

type testTask struct {
}

func (t *testTask) Run() {
	fmt.Println("hello world 1")
}

func SendMail(mailTo []string, subject string, body string) error {
	//定义邮箱服务器连接信息，如果是阿里邮箱 pass填密码，qq邮箱填授权码
	mailConn := map[string]string{
		"user": "zibianqu@163.com",
		"pass": "GYMVTVAKCORWOXKR",
		"host": "smtp.163.com",
		"port": "465",
	}

	port, _ := strconv.Atoi(mailConn["port"]) //转换端口类型为int

	m := gomail.NewMessage()
	m.SetHeader("From", "自编曲"+"<"+mailConn["user"]+">") //这种方式可以添加别名，即“XD Game”， 也可以直接用<code>m.SetHeader("From",mailConn["user"])</code> 读者可以自行实验下效果
	m.SetHeader("To", mailTo...)                        //发送给多个用户
	m.SetHeader("Subject", subject)                     //设置邮件主题
	m.SetBody("text/html", body)                        //设置邮件正文

	d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])

	err := d.DialAndSend(m)
	return err

}

func main() {
	crontab := NewCrontab()
	// 实现接口的方式添加定时任务
	task := &testTask{}
	if err := crontab.AddJobByInterface("1", "*/1 * * * * ?", task); err != nil {
		fmt.Printf("error to add crontab task:%s", err)
		os.Exit(-1)
	}

	// 添加函数作为定时任务
	taskFunc := func() {
		fmt.Println("hello world 2")
		crontab.DelByID("1")
		//删除id为1的定时任务
		mailTo := []string{"1138675081@qq.com"}
		SendMail(mailTo, "测试定时发送", "测试定时发送")
	}
	if err := crontab.AddJobByFunc("2", "*/10 * * * * ? ", taskFunc); err != nil {
		fmt.Printf("error to add crontab task:%s", err)
		os.Exit(-1)
	}
	crontab.Start()
	select {}

}
