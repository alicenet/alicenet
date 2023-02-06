package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func overWriteUponUserCall(logger *logrus.Logger, options []*cobra.Command) {

	/* The logic here feels backwards to me but it isn't.
	Command line flags aren't set till this func returns.
	So we set flags from config file here and when func returns the command line will overwrite.
	*/
	for _, cmd := range options {
		// Find all the flags
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {

			// -help defined by pflag internals and will not parse correctly.
			if flag.Name == "help" {
				return
			}
			value := viper.Get(flag.Name)
			var err error
			if value != nil {
				datatype := reflect.TypeOf(value).String()
				flagExistingValue := flag.Value
				logger.Infof("flag name %d : flag value %d: datatype : %d", flag.Name, value, datatype)
				if datatype == "string" {
					if flagExistingValue.String() == "" {
						err = flag.Value.Set(value.(string))
					}
				} else if datatype == "bool" {
					if flagExistingValue.String() == "false" {
						err = flag.Value.Set(strconv.FormatBool(value.(bool)))
					}
				} else if datatype == "int64" {
					err = flag.Value.Set(strconv.FormatInt(value.(int64), 10))
				} else if datatype == "uint64" {
					logger.Infof("interface type is %s:  name %s : ", flag.Value.Type(), flag.Name)
					logger.Infof("interface type is %s: value  %s name %s", flag.Value.Type(), value.(uint64), flag.Name)
					if flag.Value.String() != "0" {
						err = flag.Value.Set(strconv.FormatUint(value.(uint64), 10))
					}
				} else if datatype == "duration" {
					duration := value.(time.Duration)
					s := duration.String()
					if strings.HasSuffix(s, "m0s") {
						s = s[:len(s)-2]
					}
					if strings.HasSuffix(s, "h0m") {
						s = s[:len(s)-2]
					}
					err = flag.Value.Set(s)
				}
			}

			if err != nil {
				logger.Warnf("Setting flag %q failed:%q", flag.Name, err)
			}
		})
	}
}
