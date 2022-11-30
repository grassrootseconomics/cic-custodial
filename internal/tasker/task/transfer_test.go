package task

import (
	"math/big"
	"reflect"
	"testing"
)

func Test_parseTransferValue(t *testing.T) {
	type args struct {
		value         string
		tokenDecimals int
	}
	tests := []struct {
		name string
		args args
		want *big.Int
	}{
		{
			name: "zero value string",
			args: args{
				value:         "0",
				tokenDecimals: 6,
			},
			want: big.NewInt(0),
		},
		{
			name: "fixed value string",
			args: args{
				value:         "2",
				tokenDecimals: 6,
			},
			want: big.NewInt(2000000),
		},
		{
			name: "float (2 d.p) value string",
			args: args{
				value:         "2.19",
				tokenDecimals: 6,
			},
			want: big.NewInt(2190000),
		},
		{
			name: "float (6 d.p) value string",
			args: args{
				value:         "2.123456",
				tokenDecimals: 6,
			},
			want: big.NewInt(2123456),
		},
		{
			name: "float (10 d.p) value string",
			args: args{
				value:         "2.1234567891",
				tokenDecimals: 6,
			},
			want: big.NewInt(2123456),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTransferValue(tt.args.value, tt.args.tokenDecimals)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
