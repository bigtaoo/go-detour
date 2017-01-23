package recast

import (
	"github.com/aurelien-rainone/assertgo"
	"github.com/aurelien-rainone/gogeo/f32/d3"
)

/// @par
///
/// Basically, any spans that are closer to a boundary or obstruction than the specified radius
/// are marked as unwalkable.
///
/// This method is usually called immediately after the heightfield has been built.
///
/// @see rcCompactHeightfield, rcBuildCompactHeightfield, rcConfig::walkableRadius
func ErodeWalkableArea(ctx *Context, radius int32, chf *CompactHeightfield) bool {
	assert.True(ctx != nil, "ctx should not be nil")

	w := chf.Width
	h := chf.Height

	ctx.StartTimer(RC_TIMER_ERODE_AREA)
	defer ctx.StopTimer(RC_TIMER_ERODE_AREA)

	dist := make([]uint8, chf.SpanCount)

	// Init distance.
	for i := range dist {
		dist[i] = 0xff
	}

	// Mark boundary cells.
	for y := int32(0); y < h; y++ {
		for x := int32(0); x < w; x++ {
			c := chf.Cells[x+y*w]
			ni := int32(c.Index) + int32(c.Count)
			for i := int32(c.Index); i < ni; i++ {
				if chf.Areas[i] == RC_NULL_AREA {
					dist[i] = 0
				} else {
					s := chf.Spans[i]
					nc := int32(0)
					for dir := int32(0); dir < 4; dir++ {
						if GetCon(s, dir) != RC_NOT_CONNECTED {
							nx := x + GetDirOffsetX(dir)
							ny := y + GetDirOffsetY(dir)
							nidx := int32(chf.Cells[nx+ny*w].Index) + GetCon(s, dir)
							if chf.Areas[nidx] != RC_NULL_AREA {
								nc++
							}
						}
					}
					// At least one missing neighbour.
					if nc != 4 {
						dist[i] = 0
					}
				}
			}
		}
	}

	var nd uint8

	// Pass 1
	for y := int32(0); y < h; y++ {
		for x := int32(0); x < w; x++ {
			c := chf.Cells[x+y*w]
			ni := int32(c.Index) + int32(c.Count)
			for i := int32(c.Index); i < ni; i++ {
				s := chf.Spans[i]

				if GetCon(s, 0) != RC_NOT_CONNECTED {
					// (-1,0)
					ax := x + GetDirOffsetX(0)
					ay := y + GetDirOffsetY(0)
					ai := int32(chf.Cells[ax+ay*w].Index) + int32(GetCon(s, 0))
					as := chf.Spans[ai]
					nd = uint8(iMin(int32(dist[ai]+2), int32(255)))
					if nd < dist[i] {
						dist[i] = nd
					}

					// (-1,-1)
					if GetCon(as, 3) != RC_NOT_CONNECTED {
						aax := ax + GetDirOffsetX(3)
						aay := ay + GetDirOffsetY(3)
						aai := int32(chf.Cells[aax+aay*w].Index) + int32(GetCon(as, 3))
						nd = uint8(iMin(int32(dist[aai]+3), int32(255)))
						if nd < dist[i] {
							dist[i] = nd
						}
					}
				}

				if GetCon(s, 3) != RC_NOT_CONNECTED {
					// (0,-1)
					ax := x + GetDirOffsetX(3)
					ay := y + GetDirOffsetY(3)
					ai := int32(chf.Cells[ax+ay*w].Index) + GetCon(s, 3)
					as := chf.Spans[ai]
					nd = uint8(iMin(int32(dist[ai]+2), int32(255)))
					if nd < dist[i] {
						dist[i] = nd
					}

					// (1,-1)
					if GetCon(as, 2) != RC_NOT_CONNECTED {
						aax := ax + GetDirOffsetX(2)
						aay := ay + GetDirOffsetY(2)
						aai := int32(chf.Cells[aax+aay*w].Index) + GetCon(as, 2)
						nd = uint8(iMin(int32(dist[aai]+3), int32(255)))
						if nd < dist[i] {
							dist[i] = nd
						}
					}
				}
			}
		}
	}

	// Pass 2
	for y := int32(h - 1); y >= 0; y-- {
		for x := int32(w - 1); x >= 0; x-- {
			c := chf.Cells[x+y*w]
			i := int32(c.Index)
			for ni := int32(c.Index) + int32(c.Count); i < ni; i++ {
				s := chf.Spans[i]

				if GetCon(s, 2) != RC_NOT_CONNECTED {
					// (1,0)
					ax := x + GetDirOffsetX(2)
					ay := y + GetDirOffsetY(2)
					ai := int32(chf.Cells[ax+ay*w].Index) + GetCon(s, 2)
					as := chf.Spans[ai]
					nd = uint8(iMin(int32(dist[ai]+2), int32(255)))
					if nd < dist[i] {
						dist[i] = nd
					}

					// (1,1)
					if GetCon(as, 1) != RC_NOT_CONNECTED {
						aax := ax + GetDirOffsetX(1)
						aay := ay + GetDirOffsetY(1)
						aai := int32(chf.Cells[aax+aay*w].Index) + GetCon(as, 1)
						nd = uint8(iMin(int32(dist[aai]+3), int32(255)))
						if nd < dist[i] {
							dist[i] = nd
						}
					}
				}
				if GetCon(s, 1) != RC_NOT_CONNECTED {
					// (0,1)
					ax := x + GetDirOffsetX(1)
					ay := y + GetDirOffsetY(1)
					ai := int32(chf.Cells[ax+ay*w].Index) + GetCon(s, 1)
					as := chf.Spans[ai]
					nd = uint8(iMin(int32(dist[ai]+2), int32(255)))
					if nd < dist[i] {
						dist[i] = nd
					}

					// (-1,1)
					if GetCon(as, 0) != RC_NOT_CONNECTED {
						aax := ax + GetDirOffsetX(0)
						aay := ay + GetDirOffsetY(0)
						aai := int32(chf.Cells[aax+aay*w].Index) + GetCon(as, 0)
						nd = uint8(iMin(int32(dist[aai]+3), int32(255)))
						if nd < dist[i] {
							dist[i] = nd
						}
					}
				}
			}
		}
	}

	thr := uint8(radius * 2)
	for i := int32(0); i < chf.SpanCount; i++ {
		if dist[i] < thr {
			chf.Areas[i] = RC_NULL_AREA
		}
	}

	dist = nil

	return true
}

/// @par
///
/// The value of spacial parameters are in world units.
///
/// The y-values of the polygon vertices are ignored. So the polygon is effectively
/// projected onto the xz-plane at @p hmin, then extruded to @p hmax.
///
/// @see rcCompactHeightfield, rcMedianFilterWalkableArea
func MarkConvexPolyArea(ctx *Context, verts []float32, nverts int32,
	hmin, hmax float32, areaId uint8, chf *CompactHeightfield) {

	assert.True(ctx != nil, "ctx should not be nil")

	ctx.StartTimer(RC_TIMER_MARK_CONVEXPOLY_AREA)
	defer ctx.StopTimer(RC_TIMER_MARK_CONVEXPOLY_AREA)

	var bmin, bmax [3]float32
	copy(bmin[:], verts[:3])
	copy(bmax[:], verts[:3])
	for i := int32(1); i*3 < nverts; i++ {
		v := verts[i*3:]
		d3.Vec3Min(bmin[:], v)
		d3.Vec3Max(bmax[:], v)
	}
	bmin[1] = hmin
	bmax[1] = hmax

	minx := int32(((bmin[0] - chf.BMin[0]) / chf.Cs))
	miny := int32(((bmin[1] - chf.BMin[1]) / chf.Ch))
	minz := int32(((bmin[2] - chf.BMin[2]) / chf.Cs))
	maxx := int32(((bmax[0] - chf.BMin[0]) / chf.Cs))
	maxy := int32(((bmax[1] - chf.BMin[1]) / chf.Ch))
	maxz := int32(((bmax[2] - chf.BMin[2]) / chf.Cs))

	if maxx < 0 {
		return
	}
	if minx >= chf.Width {
		return
	}
	if maxz < 0 {
		return
	}
	if minz >= chf.Height {
		return
	}

	if minx < 0 {
		minx = 0
	}
	if maxx >= chf.Width {
		maxx = chf.Width - 1
	}
	if minz < 0 {
		minz = 0
	}
	if maxz >= chf.Height {
		maxz = chf.Height - 1
	}

	// TODO: Optimize.
	for z := minz; z <= maxz; z++ {
		for x := minx; x <= maxx; x++ {
			c := chf.Cells[x+z*chf.Width]
			i := int32(c.Index)
			for ni := int32(c.Index) + int32(c.Count); i < ni; i++ {
				s := chf.Spans[i]
				if chf.Areas[i] == RC_NULL_AREA {
					continue
				}
				if int32(s.Y) >= miny && int32(s.Y) <= maxy {
					var p [3]float32
					p[0] = chf.BMin[0] + (float32(x)+0.5)*chf.Cs
					p[1] = 0
					p[2] = chf.BMin[2] + (float32(z)+0.5)*chf.Cs

					if pointInPoly(nverts, verts, p[:]) {
						chf.Areas[i] = areaId
					}
				}
			}
		}
	}
}

func pointInPoly(nvert int32, verts, p []float32) bool {
	var (
		i, j int32
		c    bool
	)

	// TODO: check that, j = i++
	//for j = nvert-1; i < nvert; j = i++) {
	for j = nvert - 1; i < nvert; i++ {
		vi := verts[i*3:]
		vj := verts[j*3:]
		if ((vi[2] > p[2]) != (vj[2] > p[2])) &&
			(p[0] < (vj[0]-vi[0])*(p[2]-vi[2])/(vj[2]-vi[2])+vi[0]) {
			c = !c
		}
		j = i
	}
	return c
}