package main


import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/gocolly/colly/proxy"
	"io"
	"net/url"
	"os"
	"regexp"
)

func main()  {
	var targetName string
	fmt.Println("请输入要抓取的用户名")
	fmt.Scanln(&targetName) //输入要抓取的用户名

	urls := "https://www.instagram.com/" + targetName //拼接url
	aId,aCur := getIdAndCursor(urls)
	fmt.Println(aCur,aId)

	fileNameImg := "./GoProject/MyProject/src/download"+targetName+"_img"
	fileNameVideo := "./GoProject/MyProject/src/download"+targetName+"_video"
	dirmk(fileNameImg)
	dirmk(fileNameVideo)

	v := url.Values{}
	v.Add("after",aCur)
	v.Add("first","50")
	v.Add("id",aId)
	//拼接url
	body := "https://www.instagram.com/graphql/query/?query_hash=e769aa130647d2354c40ea6a439bfc08&" + v.Encode()
	getCount(body,aId,fileNameImg,fileNameVideo)
}

func getCount(urls,id,fileNameImg,fileNameVideo string)  {
	// 获取图片以及视频的url 并下载
	var path string // 存放路径
	c := colly.NewCollector(func(collector *colly.Collector) {
		extensions.RandomUserAgent(collector)
	})

	imageC := c.Clone()

	if p,err := proxy.RoundRobinProxySwitcher(
		"http://127.0.0.1:8888",);
	err == nil{
		c.SetProxyFunc(p)
	}

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cookie", "ig_did=CDD46F6A-1B7A-43A8-8660-4C9C843FD8FE; mid=XncUvgALAAEGp32npsS73P_cbRKR; csrftoken=yCsmfUzGw1U6XNHPPOL2jGLOjs8dSth7; ds_user_id=32233184728; sessionid=32233184728%3A5lMCAovAehJdts%3A15; rur=FRC; urlgen='{\"65.49.126.82\": 6939}:1jFx1z:wXy5nuI6z2ntuF9F6pdWH7o-Zqs'")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Accept", "*/*")
	})

	c.OnResponse(func(response *colly.Response) {
		var f interface{}
		json.Unmarshal(response.Body,&f)
		fmt.Println(json.Unmarshal(response.Body,&f))
		data := f.(map[string]interface{})["data"]
		user := data.(map[string]interface{})["user"]
		edge_owner := user.(map[string]interface{})["edge_owner_to_timeline_media"]

		edges := edge_owner.(map[string]interface{})["edges"]
		for k,v := range edges.([]interface{}){
			node := v.(map[string]interface{})["node"]
			imageUrl := node.(map[string]interface{})["display_url"].(string)
			path = fileNameImg + "/"
			fmt.Println(k,imageUrl)
			imageC.Visit(imageUrl)

			if node.(map[string]interface{})["is_video"] == true{
				path = fileNameVideo + "/"
				videoUrl := node.(map[string]interface{})["video_url"].(string)
				fmt.Println(k,videoUrl)
				imageC.Visit(videoUrl)
			}else{
				fmt.Println("没有视频可以下载")
			}
		}
		page_info := edge_owner.(map[string]interface{})["page_info"]
		has_next_page := page_info.(map[string]interface{})["has_next_page"]

		if has_next_page == true{
			//如果有下一页则继续获取
			end_cursor := page_info.(map[string]interface{})["end_cursor"].(string)
			v := url.Values{}
			v.Add("after",end_cursor)
			v.Add("first","50")
			v.Add("id",id)
			body := "https://www.instagram.com/graphql/query/?query_hash=e769aa130647d2354c40ea6a439bfc08&" + v.Encode()
			//拼接url
			c.Visit(body)
			fmt.Println("body>",body)
		}
	})
	c.OnError(func(response *colly.Response, err error) {
		fmt.Println(err)
	})
	fmt.Println(path)
	imageC.OnResponse(func(r *colly.Response) {
		fileName := ""
		reg := regexp.MustCompile("(\\d+_n.*)\\?")
		caption := reg.FindAllStringSubmatch(r.Request.URL.String(),-1)

		if caption == nil{
			fmt.Println("regexp err")
			return
		}
		fileName = caption[0][1]
		fmt.Println(fileName)
		fmt.Printf("download -->%s \n",fileName)
		f,err := os.Create(path + fileName)
		if err != nil{
			panic(err)
		}
		io.Copy(f,bytes.NewReader(r.Body))
	})
	c.Visit(urls)
	c.Wait()
	imageC.Wait()
}


func getIdAndCursor(urls string) (string,string) {
	var aId string
	var aCur string

	c := colly.NewCollector(func(collector *colly.Collector) {
		extensions.RandomUserAgent(collector)
	})
	if p,err := proxy.RoundRobinProxySwitcher(
		"http://127.0.0.1:8888",
		);err == nil{
		c.SetProxyFunc(p)
	}

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cookie", "ig_did=CDD46F6A-1B7A-43A8-8660-4C9C843FD8FE; mid=XncUvgALAAEGp32npsS73P_cbRKR; csrftoken=yCsmfUzGw1U6XNHPPOL2jGLOjs8dSth7; ds_user_id=32233184728; sessionid=32233184728%3A5lMCAovAehJdts%3A15; rur=FRC; urlgen='{\"65.49.126.82\": 6939}:1jFx1z:wXy5nuI6z2ntuF9F6pdWH7o-Zqs'")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Accept", "*/*")
	}) //添加请求头和cookie

	c.OnRequest(func(request *colly.Request) {
		fmt.Printf("fetch --->%s\n",request.URL.String())
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		reg_id := regexp.MustCompile("profilePage_(\\d+)")
		reg_cur := regexp.MustCompile("\"end_cursor\":\"(.*?)\"")
		idres := reg_id.FindAllStringSubmatch(e.Text,-1)
		curres := reg_cur.FindAllStringSubmatch(e.Text,-1)
		aId = idres[0][1]
		aCur = curres[0][1]
	})
	c.Visit(urls)
	c.Wait()
	return aId,aCur

}

func dirmk(path string)  {
	//如果文件夹不存在则创建
	_dir := path
	exist,err := pathExists(_dir)
	if err != nil{
		fmt.Printf("get dir error[%v]\n",err)
		return
	}
	if exist{
		fmt.Printf("has dir[%v]\n",_dir)
	}else {
		fmt.Printf("no dir[%v]\n",_dir)
		//创建文件夹
		err := os.Mkdir(_dir,os.ModePerm)
		if err != nil{
			fmt.Printf("mkdir failed[%v]\n",err)
		}else {
			fmt.Printf("mkdir success\n")
		}
	}
}

func pathExists(path string) (bool,error) {
	//检查文件夹是否存在
	_,err := os.Stat(path)
	if err == nil{
		return true,nil
	}
	if os.IsNotExist(err){
		return false,nil
	}
	return false,err
}