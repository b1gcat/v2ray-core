package apk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ext(t *testing.T) {
	ext, err := getPathFromUrl(`GET /stage/stream-403255979171381930_or4.flv?302_type=cold_aggr&_session_id=552-202404251206362813060339A8BD71935C.1714017996281.60837&abr_pts=-800&align_backward=true&align_delay=-35&cb_retry=0&domain=pull-hs-f5.flive.douyincdn.com&expire=1714622786&fp_user_url=https%3A%2F%2Fpull-hs-f5.flive.douyincdn.com%2Fstage%2Fstream-403255979171381930_or4.flv%3Fexpire%3D1714622786%26sign%3Dfc56e781fb186632db8716aed55230ce%26volcSecret%3Dfc56e781fb186632db8716aed55230ce%26volcTime%3D1714622786%26abr_pts%3D-800%26_session_id%3D552-202404251206362813060339A8BD71935C.1714017996281.60837&manage_ip=&node_id=&pro_type=http2&redirect_from=pod.cn-fc968z.otc6.nss&sign=fc56e781fb186632db8716aed55230ce&vhost=push-rtmp-hs-f5.douyincdn.com&volcSecret=fc56e781fb186632db8716aed55230ce&volcTime=1714622786 HTTP1.1`, 4)
	assert.Equal(t, nil, err)
	assert.Equal(t, `/stage/stream-403255979171381930_or4.flv?302_type=cold_aggr&_session_id=552-202404251206362813060339A8BD71935C.1714017996281.60837&abr_pts=-800&align_backward=true&align_delay=-35&cb_retry=0&domain=pull-hs-f5.flive.douyincdn.com&expire=1714622786&fp_user_url=https%3A%2F%2Fpull-hs-f5.flive.douyincdn.com%2Fstage%2Fstream-403255979171381930_or4.flv%3Fexpire%3D1714622786%26sign%3Dfc56e781fb186632db8716aed55230ce%26volcSecret%3Dfc56e781fb186632db8716aed55230ce%26volcTime%3D1714622786%26abr_pts%3D-800%26_session_id%3D552-202404251206362813060339A8BD71935C.1714017996281.60837&manage_ip=&node_id=&pro_type=http2&redirect_from=pod.cn-fc968z.otc6.nss&sign=fc56e781fb186632db8716aed55230ce&vhost=push-rtmp-hs-f5.douyincdn.com&volcSecret=fc56e781fb186632db8716aed55230ce&volcTime=1714622786`, ext)
}
