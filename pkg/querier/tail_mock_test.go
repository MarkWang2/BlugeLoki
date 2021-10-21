package querier

import "github.com/MarkWang2/BlugeLoki/pkg/logproto"

func mockTailResponse(stream logproto.Stream) *logproto.TailResponse {
	return &logproto.TailResponse{
		Stream:         &stream,
		DroppedStreams: []*logproto.DroppedStream{},
	}
}
