package execute

// 替换为你的服务器信息
const (
	IP       = "192.168.210.197"
	Port     = 60800
	User     = "hc"
	Password = "geotellHC666"
	LOD2Path = "/home/hc/LOD2exe/"
)

// Config 定义了默认参数配置
var DefaultConfig = map[string]string{
	"data_path":         LOD2Path + "Data/obj_data/original_tile",
	"save_path":         LOD2Path + "Data/obj_data/images",
	"log_name":          LOD2Path + "Data/obj_data/Logs/",
	"color_path":        LOD2Path + "configSL/colorname.txt",
	"threads":           "48",
	"scale":             "0.1",
	"labelnum":          "13",
	"bRenderImage":      "1",
	"bBackProjectImage": "0",
}

// RequestParams 定义接收前端 JSON 请求的结构体
type RequestParams struct {
	DataPath          string `json:"data_path,omitempty"`
	SavePath          string `json:"save_path,omitempty"`
	LogName           string `json:"log_name,omitempty"`
	ColorPath         string `json:"color_path,omitempty"`
	Threads           string `json:"threads,omitempty"`
	Scale             string `json:"scale,omitempty"`
	LabelNum          string `json:"labelnum,omitempty"`
	BRenderImage      string `json:"bRenderImage,omitempty"`
	BBackProjectImage string `json:"bBackProjectImage,omitempty"`
}
