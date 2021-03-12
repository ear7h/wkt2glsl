package main

import (
	"math"
	"context"
	"fmt"
	"os"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/simplify"
)

func PointCount(geo geom.Geometry) int {
	switch v := geo.(type) {
	default:
		panic(fmt.Sprintf("unknown geometry %T", geo))

	case geom.Point:
		return 1

	case geom.MultiPoint:
		return len(v)

	case geom.LineString:
		return len(v)

	case geom.MultiLineString:
		sum := 0
		for _, vv := range v {
			sum += PointCount(geom.LineString(vv))
		}
		return sum

	case geom.Polygon:
		sum := 0
		for _, vv := range v {
			sum += PointCount(geom.LineString(vv))
		}
		return sum

	case geom.MultiPolygon:
		sum := 0
		for _, vv := range v {
			sum += PointCount(geom.Polygon(vv))
		}
		return sum
	}
}

func RemoveEmpty(geo geom.Geometry) geom.Geometry {
	switch v := geo.(type) {
	default:
		panic(fmt.Sprintf("unknown geometry %T", geo))

	case geom.Point:
		if v != v {
			return nil
		}
		return v

	case geom.MultiPoint:
		if len(v) == 0 {
			return nil
		}

		ret := make([][2]float64, 0, len(v))
		for _, vv := range v {
			tmp := RemoveEmpty(geom.Point(vv))
			if tmp != nil {
				ret = append(ret, tmp.(geom.Point))
			}
		}

		if len(ret) == 0 {
			return nil
		}

		return geom.MultiPoint(ret)

	case geom.LineString:
		if len(v) == 0 {
			return nil
		}

		ret := make([][2]float64, 0, len(v))
		for _, vv := range v {
			tmp := RemoveEmpty(geom.Point(vv))
			if tmp != nil {
				ret = append(ret, tmp.(geom.Point))
			}
		}

		if len(ret) == 0 {
			return nil
		}

		return geom.LineString(ret)

	case geom.MultiLineString:
		if len(v) == 0 {
			return nil
		}
		ret := make([][][2]float64, 0, len(v))
		for _, vv := range v {
			tmp := RemoveEmpty(geom.LineString(vv))
			if tmp != nil {
				ret = append(ret, tmp.(geom.LineString))
			}
		}

		if len(ret) == 0 {
			return nil
		}

		return geom.MultiLineString(ret)

	case geom.Polygon:
		if len(v) == 0 {
			return nil
		}

		ret := make([][][2]float64, 0, len(v))
		for _, vv := range v {
			tmp := RemoveEmpty(geom.LineString(vv))
			if tmp != nil {
				ret = append(ret, tmp.(geom.LineString))
			}
		}

		if len(ret) == 0 {
			return nil
		}

		return geom.Polygon(ret)

	case geom.MultiPolygon:
		if len(v) == 0 {
			return nil
		}

		ret := make([][][][2]float64, 0, len(v))
		for _, vv := range v {
			tmp := RemoveEmpty(geom.Polygon(vv))
			if tmp != nil {
				ret = append(ret, tmp.(geom.Polygon))
			}
		}

		if len(ret) == 0 {
			return nil
		}

		return geom.MultiPolygon(ret)
	}
}

func main() {
	geo, err := wkt.Decode(os.Stdin)
	if err != nil {
		panic(err)
	}

	geo = flattenCollection(geo.(geom.Collection))

	fmt.Fprintln(os.Stderr, PointCount(geo))

	geo, err = planar.Simplify(context.Background(), simplify.DouglasPeucker{
		Tolerance: 6.0,
	}, geo)
	if err != nil {
		panic(err)
	}


	geo = RemoveEmpty(geo)
	geo = removeSmall(geo.(geom.MultiLineString))

	fmt.Fprintln(os.Stderr, PointCount(geo))

	/*
	err = wkt.NewEncoder(os.Stderr, false, 10, 'g').Encode(geo)
	if err != nil {
		panic(err)
	}
	*/

	printVecs(geo.(geom.MultiLineString))
}

func flattenCollection(geo geom.Collection) geom.MultiLineString {
	n := 0
	for _, v := range geo {
		n += len(v.(geom.MultiLineString))
	}

	ret := make([][][2]float64, 0, n)
	for _, v := range geo {
		ret = append(ret, v.(geom.MultiLineString)...)
	}
	return ret
}

func removeSmall(geo geom.MultiLineString) geom.MultiLineString {
	ret := make([][][2]float64, 0, len(geo))
	for _, v := range geo {
		if len(v) > 2 {
			ret = append(ret, v)
		}
	}

	return ret
}

func printVecs(geo geom.MultiLineString) {
	for _, line := range geo {
		for _, point := range line {
			fmt.Printf("vec2(%f, %f),\n",
			deg2rad(point[0]),
			deg2rad(point[1]))
		}
		fmt.Printf("vec2(0., 0.),\n")
	}
}

func deg2rad(d float64) float64 {
	return math.Pi * d / 180
}
