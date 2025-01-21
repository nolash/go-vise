//go:build !logwarn && !logdebug && !loginfo && !logtrace && !logerror
// +build !logwarn,!logdebug,!loginfo,!logtrace,!logerror

package logging

var (
	LogLevel = LVL_NONE
)
