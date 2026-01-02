package plextv

import (
	"encoding/xml"
	"testing"
	"time"
)

func TestPlexTimestamp_UnmarshalXMLAttr(t *testing.T) {
	tests := []struct {
		name    string
		attr    xml.Attr
		want    time.Time
		wantErr bool
	}{
		{
			name: "valid timestamp",
			attr: xml.Attr{
				Name:  xml.Name{Local: "createdAt"},
				Value: "1609459200", // 2021-01-01 00:00:00 UTC
			},
			want:    time.Unix(1609459200, 0),
			wantErr: false,
		},
		{
			name: "another valid timestamp",
			attr: xml.Attr{
				Name:  xml.Name{Local: "lastSeenAt"},
				Value: "1700000000",
			},
			want:    time.Unix(1700000000, 0),
			wantErr: false,
		},
		{
			name: "invalid timestamp (non-numeric)",
			attr: xml.Attr{
				Name:  xml.Name{Local: "createdAt"},
				Value: "abc",
			},
			wantErr: true,
		},
		{
			name: "empty timestamp",
			attr: xml.Attr{
				Name:  xml.Name{Local: "createdAt"},
				Value: "",
			},
			wantErr: true,
		},
		{
			name: "negative timestamp",
			attr: xml.Attr{
				Name:  xml.Name{Local: "createdAt"},
				Value: "-62135596800", // time.Unix(-62135596800, 0) is roughly year 0
			},
			want:    time.Unix(-62135596800, 0),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pt PlexTimestamp
			err := pt.UnmarshalXMLAttr(tt.attr)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalXMLAttr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !time.Time(pt).Equal(tt.want) {
					t.Errorf("UnmarshalXMLAttr() got = %v, want %v", time.Time(pt), tt.want)
				}
			}
		})
	}
}
