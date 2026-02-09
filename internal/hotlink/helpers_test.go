package hotlink_test

import (
	"testing"

	"github.com/ailinykh/reposter/v3/internal/hotlink"
)

func Test_HotlinkExpandURL(t *testing.T) {
	testCases := []struct {
		name string
		when string
		then string
	}{
		{
			name: "it handles a regular twitter post url",
			when: "https://twitter.com/username/status/123456",
			then: "https://x.com/status/123456",
		},
		{
			name: "it handles a link to a x.com photo",
			when: "https://x.com/x/status/654321/photo/1",
			then: "https://x.com/status/654321",
		},
		{
			name: "it ignores non-twitter hostnames",
			when: "https://notx.com/x/username/status/123456",
			then: "",
		},
		{
			name: "it handles a regulat youtube url",
			when: "https://www.youtube.com/watch?v=JfbnpYLe3Ms",
			then: "https://youtu.be/JfbnpYLe3Ms",
		},
		{
			name: "it handles a youtu.be format",
			when: "https://youtu.be/JfbnpYLe3Ms",
			then: "https://youtu.be/JfbnpYLe3Ms",
		},
		{
			name: "it handles a youtube shorts",
			when: "https://youtube.com/shorts/JfbnpYLe3Ms",
			then: "https://youtu.be/JfbnpYLe3Ms",
		},
		{
			name: "it ignores youtube channel link",
			when: "https://www.youtube.com/channel/UCav4Xyxlt8XX23Oz96QAJPA",
			then: "",
		},
		{
			name: "it ignores short youtube channel link",
			when: "https://www.youtube.com/@DanielLaBelle",
			then: "",
		},
		{
			name: "it ignores youtube playlist",
			when: "https://www.youtube.com/watch?v=playlist&feature=youtu.be",
			then: "",
		},
		{
			name: "it ignores youtube clip",
			when: "https://www.youtube.com/watch?v=clip&feature=youtu.be",
			then: "",
		},
		{
			name: "it ignores youtube embed",
			when: "https://www.youtube.com/watch?v=embed&feature=youtu.be",
			then: "",
		},
		{
			name: "it handles an instagram reel",
			when: "https://www.instagram.com/reel/DNVr-n8puEO/?hl=en",
			then: "https://instagram.com/reel/DNVr-n8puEO",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := hotlink.ExpandURL(testCase.when)

			if result != testCase.then {
				t.Errorf(`expected "%s", but got "%s"`, testCase.then, result)
			}

			if testCase.then == "" && err == nil {
				t.Error("expected error to exist, got nil instead")
			}
		})
	}
}
