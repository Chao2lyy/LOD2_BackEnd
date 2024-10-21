package execute

import (
	"encoding/xml"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// jsonToXML 将 JSON 数据转换为 XML 格式
func jsonToXML(jsonData map[string]interface{}) (string, error) {
	// 填充配置结构体
	config := Config{}

	// 从 JSON 数据中获取并设置配置项
	if prefix, ok := jsonData["prefix"].(string); ok {
		config.Prefix = prefix
	} else {
		return "", errors.New("prefix 字段缺失")
	}
	config.InputDir = filepath.ToSlash(filepath.Join(LOD2Path, "input_temp")) + "/"
	config.OutputDir = filepath.ToSlash(filepath.Join(LOD2Path, "output_temp")) + "/"

	if recThreadNum, ok := jsonData["rec_thread_num"].(float64); ok {
		config.RecThreadNum = int(recThreadNum)
	} else {
		config.RecThreadNum = 4 // 默认值
	}

	if calThreadNum, ok := jsonData["cal_thread_num"].(float64); ok {
		config.CalThreadNum = int(calThreadNum)
	} else {
		config.CalThreadNum = 2 // 默认值
	}

	// Scene 部分
	if sceneData, ok := jsonData["scene"].(map[string]interface{}); ok {
		if step, ok := sceneData["step"].(float64); ok {
			config.Scene.Step = step
		} else {
			return "", errors.New("scene 中 step 字段缺失")
		}
	}

	// Block 部分
	if blockData, ok := jsonData["block"].(map[string]interface{}); ok {
		if height, ok := blockData["height"].(float64); ok {
			config.Block.Height = int(height)
		}
		if width, ok := blockData["width"].(float64); ok {
			config.Block.Width = int(width)
		}
		if writeData, ok := blockData["write"].(map[string]interface{}); ok {
			config.Block.Write.HeightMap = writeData["height_map"].(bool)
			config.Block.Write.NormalMap = writeData["normal_map"].(bool)
		}
	}

	// Building 部分
	if buildingData, ok := jsonData["building"].(map[string]interface{}); ok {
		if minArea, ok := buildingData["min_area"].(float64); ok {
			config.Building.MinArea = minArea
		}

		// 处理 VSA
		if vsaData, ok := buildingData["vsa"].(map[string]interface{}); ok {
			if seedArea, ok := vsaData["seed_area"].(float64); ok {
				config.Building.VSA.SeedArea = seedArea
			}
			if iterations, ok := vsaData["iterations"].(float64); ok {
				config.Building.VSA.Iterations = int(iterations)
			}
		}

		// 处理 FacadeSegment
		if facadeSegmentData, ok := buildingData["facade_segment"].(map[string]interface{}); ok {
			if vsa, ok := facadeSegmentData["vsa"].(map[string]interface{}); ok {
				if seedArea, ok := vsa["seed_area"].(float64); ok {
					config.Building.FacadeSegment.VSA.SeedArea = seedArea
				}
				if iterations, ok := vsa["iterations"].(float64); ok {
					config.Building.FacadeSegment.VSA.Iterations = int(iterations)
				}
			}
			if minArea, ok := facadeSegmentData["min_area"].(float64); ok {
				config.Building.FacadeSegment.MinArea = minArea
			}
			if normalThreshold, ok := facadeSegmentData["normal_threshold"].(float64); ok {
				config.Building.FacadeSegment.NormalThreshold = normalThreshold
			}
		}
		fmt.Println("buid", buildingData["image"])
		// 处理 Image 部分
		if imageData, ok := buildingData["image"].(map[string]interface{}); ok {
			if lsdData, ok := imageData["lsd"].(map[string]interface{}); ok {
				config.Building.Image.LSD = parseLSD(lsdData)
			}

		}

		// 处理 Depth 部分
		fmt.Println("buid", buildingData["depth"])
		if depthData, ok := buildingData["depth"].(map[string]interface{}); ok {
			if lsdData, ok := depthData["lsd"].(map[string]interface{}); ok {
				config.Building.Depth.LSD = parseLSD(lsdData)
			}
		}

		// 处理 Normal 部分
		if normalData, ok := buildingData["normal"].(map[string]interface{}); ok {
			if lsdData, ok := normalData["lsd"].(map[string]interface{}); ok {
				config.Building.Normal.LSD = parseLSD(lsdData)
			}
		}

		// 处理 FacadeDirection 部分
		if facadeDirectionData, ok := buildingData["facade_direction"].(map[string]interface{}); ok {
			if minArea, ok := facadeDirectionData["min_area"].(float64); ok {
				config.Building.FacadeDirection.MinArea = minArea
			}
			if normalThreshold, ok := facadeDirectionData["normal_threshold"].(float64); ok {
				config.Building.FacadeDirection.NormalThreshold = normalThreshold
			}
			if verticalTolerance, ok := facadeDirectionData["vertical_tolerance"].(float64); ok {
				config.Building.FacadeDirection.VerticalTolerance = int(verticalTolerance)
			}
			if ransacData, ok := facadeDirectionData["ransac"].(map[string]interface{}); ok {
				config.Building.FacadeDirection.Ransac = parseRansac(ransacData)
			}
		}

		// 处理 RoofDirection 部分
		if roofDirectionData, ok := buildingData["roof_direction"].(map[string]interface{}); ok {
			if ransacData, ok := roofDirectionData["ransac"].(map[string]interface{}); ok {
				config.Building.RoofDirection.Ransac = parseRansac(ransacData)
			}
		}

		// 处理 PlaneDetection 部分
		if planeDetectionData, ok := buildingData["plane_detection"].(map[string]interface{}); ok {
			if ransacData, ok := planeDetectionData["ransac"].(map[string]interface{}); ok {
				config.Building.PlaneDetection.Ransac = parseRansac(ransacData)
			}
		}

		// 处理 Regularization 部分
		if regularizationData, ok := buildingData["regularization"].(map[string]interface{}); ok {
			if parallelThreshold, ok := regularizationData["parallel_threshold"].(float64); ok {
				config.Building.Regularization.ParallelThreshold = int(parallelThreshold)
			}
			if collinearThreshold, ok := regularizationData["collinear_threshold"].(float64); ok {
				config.Building.Regularization.CollinearThreshold = collinearThreshold
			}
			if parallelThresholdRe, ok := regularizationData["parallel_threshold_re"].(float64); ok {
				config.Building.Regularization.ParallelThresholdRe = int(parallelThresholdRe)
			}
			if collinearDisThreshold, ok := regularizationData["collinear_dis_threshold"].(float64); ok {
				config.Building.Regularization.CollinearDisThreshold = collinearDisThreshold
			}
		}

		// 处理 ARR 部分
		if arrData, ok := buildingData["arr"].(map[string]interface{}); ok {
			if extendRatio, ok := arrData["extend_ratio"].(float64); ok {
				config.Building.ARR.ExtendRatio = extendRatio
			}
			if mrfData, ok := arrData["mrf"].(map[string]interface{}); ok {
				if useSwap, ok := mrfData["use_swap"].(bool); ok {
					config.Building.ARR.MRF.UseSwap = useSwap
				}
				if balance, ok := mrfData["balance"].(float64); ok {
					config.Building.ARR.MRF.Balance = balance
				}
				if iterations, ok := mrfData["iterations"].(float64); ok {
					config.Building.ARR.MRF.Iterations = int(iterations)
				}
			}
		}

		// 处理 BuildingWrite 部分
		if writeData, ok := buildingData["write"].(map[string]interface{}); ok {
			config.Building.Write.ColorSegments = writeData["color_segments"].(bool)
			config.Building.Write.HeightSegments = writeData["height_segments"].(bool)
			config.Building.Write.NormalSegments = writeData["normal_segments"].(bool)
			config.Building.Write.FacadeSegments = writeData["facade_segments"].(bool)
			config.Building.Write.AllSegments = writeData["all_segments"].(bool)
			config.Building.Write.RegularizedSegments = writeData["regularized_segments"].(bool)
			config.Building.Write.Arrangement = writeData["arrangement"].(bool)
			config.Building.Write.ArrangementLabeling = writeData["arrangement_labeling"].(bool)
			config.Building.Write.LocalGround = writeData["local_ground"].(bool)
			config.Building.Write.Levels = writeData["levels"].(bool)
		}

		// 处理 Save 部分
		if saveData, ok := buildingData["save"].(map[string]interface{}); ok {
			config.Building.Save.Mesh = saveData["mesh"].(bool)
			config.Building.Save.PointCloud = saveData["point_cloud"].(bool)
		}
	}

	// 将配置转换为 XML 格式
	xmlData, err := xml.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}

	// 返回 XML 字符串
	xmlContent := xml.Header + string(xmlData)
	xmlContent = strings.Replace(xmlContent, "<Config>", "", 1)
	xmlContent = strings.Replace(xmlContent, "</Config>", "", 1)

	return xmlContent, nil
}

// 解析 LSD 子结构
func parseLSD(data map[string]interface{}) LSD {
	//fmt.Println(data["scale"])
	return LSD{
		Scale:      data["scale"].(float64),
		SigmaScale: data["sigma_scale"].(float64),
		Quant:      data["quant"].(float64),
		AngTh:      data["ang_th"].(float64),
		LogEps:     data["log_eps"].(float64),
		DensityTh:  data["density_th"].(float64),
		NBins:      int(data["n_bins"].(float64)),
	}
}

// 解析 Ransac 子结构
func parseRansac(data map[string]interface{}) Ransac {
	return Ransac{
		MinArea:         data["min_area"].(float64),
		Epsilon:         data["epsilon"].(float64),
		NormalThreshold: data["normal_threshold"].(float64),
		ClusterEpsilon:  data["cluster_epsilon"].(float64),
		Probability:     data["probability"].(float64),
	}
}
