/*
 Licensed to the Apache Software Foundation (ASF) under one
 or more contributor license agreements.  See the NOTICE file
 distributed with this work for additional information
 regarding copyright ownership.  The ASF licenses this file
 to you under the Apache License, Version 2.0 (the
 "License"); you may not use this file except in compliance
 with the License.  You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package utils

import (
	"go.uber.org/zap"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type Chart struct {
	Title      string
	XLabel     string
	YLabel     string
	Width      vg.Length
	Height     vg.Length
	LinePoints []interface{}
	SvgFile    string
}

func DrawChart(chart *Chart) error {
	p := plot.New()
	p.Title.Text = chart.Title
	p.X.Label.Text = chart.XLabel
	p.Y.Label.Text = chart.YLabel
	err := plotutil.AddLinePoints(p, chart.LinePoints...)
	if err != nil {
		return err
	}
	if err := p.Save(chart.Width, chart.Height, chart.SvgFile); err != nil {
		return err
	}
	Logger.Info("Successfully draw chart", zap.String("title", chart.Title),
		zap.String("outputFile", chart.SvgFile))
	return nil
}

func GetPointsFromSlice(slice []int) plotter.XYs {
	pts := make(plotter.XYs, len(slice))
	for i := range slice {
		pts[i].X = float64(i)
		pts[i].Y = float64(slice[i])
	}
	return pts
}

func GetLinePoints(dataMap map[string][]int) []interface{} {
	var linePoints []interface{}
	for k, v := range dataMap {
		linePoints = append(linePoints, k, GetPointsFromSlice(v))
	}
	return linePoints
}
