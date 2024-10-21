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
type Config struct {
	Prefix       string   `xml:"prefix"`
	InputDir     string   `xml:"input_dir"`
	OutputDir    string   `xml:"output_dir"`
	RecThreadNum int      `xml:"rec_thread_num"`
	CalThreadNum int      `xml:"cal_thread_num"`
	Scene        Scene    `xml:"scene"`
	Block        Block    `xml:"block"`
	Building     Building `xml:"building"`
}

type Scene struct {
	Step float64 `xml:"step"`
}

type Block struct {
	Height int   `xml:"height"`
	Width  int   `xml:"width"`
	Write  Write `xml:"write"`
}

type Write struct {
	HeightMap bool `xml:"height_map"`
	NormalMap bool `xml:"normal_map"`
}

type Building struct {
	MinArea         float64         `xml:"min_area"`
	VSA             VSA             `xml:"vsa"`
	FacadeSegment   FacadeSegment   `xml:"facade_segment"`
	Image           Image           `xml:"image"`
	Depth           Depth           `xml:"depth"`
	Normal          Normal          `xml:"normal"`
	FacadeDirection FacadeDirection `xml:"facade_direction"`
	RoofDirection   RoofDirection   `xml:"roof_direction"`
	PlaneDetection  PlaneDetection  `xml:"plane_detection"`
	Regularization  Regularization  `xml:"regularization"`
	ARR             ARR             `xml:"arr"`
	Write           BuildingWrite   `xml:"write"`
	Save            Save            `xml:"save"`
}

type VSA struct {
	SeedArea   float64 `xml:"seed_area"`
	Iterations int     `xml:"iterations"`
}

type FacadeSegment struct {
	VSA             VSA     `xml:"vsa"`
	MinArea         float64 `xml:"min_area"`
	NormalThreshold float64 `xml:"normal_threshold"`
}

type Image struct {
	LSD LSD `xml:"lsd"`
}

type Depth struct {
	LSD LSD `xml:"lsd"`
}

type Normal struct {
	LSD LSD `xml:"lsd"`
}

type LSD struct {
	Scale      float64 `xml:"scale"`
	SigmaScale float64 `xml:"sigma_scale"`
	Quant      float64 `xml:"quant"`
	AngTh      float64 `xml:"ang_th"`
	LogEps     float64 `xml:"log_eps"`
	DensityTh  float64 `xml:"density_th"`
	NBins      int     `xml:"n_bins"`
}

type FacadeDirection struct {
	MinArea           float64 `xml:"min_area"`
	NormalThreshold   float64 `xml:"normal_threshold"`
	VerticalTolerance int     `xml:"vertical_tolerance"`
	Ransac            Ransac  `xml:"ransac"`
}

type Ransac struct {
	MinArea         float64 `xml:"min_area"`
	Epsilon         float64 `xml:"epsilon"`
	NormalThreshold float64 `xml:"normal_threshold"`
	ClusterEpsilon  float64 `xml:"cluster_epsilon"`
	Probability     float64 `xml:"probability"`
}

type RoofDirection struct {
	Ransac Ransac `xml:"ransac"`
}

type PlaneDetection struct {
	Ransac Ransac `xml:"ransac"`
}

type Regularization struct {
	ParallelThreshold     int     `xml:"parallel_threshold"`
	CollinearThreshold    float64 `xml:"collinear_threshold"`
	ParallelThresholdRe   int     `xml:"parallel_threshold_re"`
	CollinearDisThreshold float64 `xml:"collinear_dis_threshold"`
}

type ARR struct {
	ExtendRatio float64 `xml:"extend_ratio"`
	MRF         MRF     `xml:"mrf"`
}

type MRF struct {
	UseSwap    bool    `xml:"use_swap"`
	Balance    float64 `xml:"balance"`
	Iterations int     `xml:"iterations"`
}

type BuildingWrite struct {
	ColorSegments       bool `xml:"color_segments"`
	HeightSegments      bool `xml:"height_segments"`
	NormalSegments      bool `xml:"normal_segments"`
	FacadeSegments      bool `xml:"facade_segments"`
	AllSegments         bool `xml:"all_segments"`
	RegularizedSegments bool `xml:"regularized_segments"`
	Arrangement         bool `xml:"arrangement"`
	ArrangementLabeling bool `xml:"arrangement_labeling"`
	LocalGround         bool `xml:"local_ground"`
	Levels              bool `xml:"levels"`
}

type Save struct {
	Mesh       bool `xml:"mesh"`
	PointCloud bool `xml:"point_cloud"`
}
