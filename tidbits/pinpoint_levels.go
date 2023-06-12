package tidbits

import (
	"cmp"
	"context"
	"log/slog"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
)

const TIDBITS_ENV_PREFIX = "TIDBITS_LOG_"

type locationPrefixEntry struct {
	LocationPrefix string
	LogLevel       slog.Level
}

type PinpointLogLevels struct {
	mtx        sync.Mutex
	levelCache map[string]slog.Level

	prefixes []locationPrefixEntry
}

func NewPinpointLogLevels() *PinpointLogLevels {
	return &PinpointLogLevels{
		prefixes:   make([]locationPrefixEntry, 0),
		levelCache: make(map[string]slog.Level),
	}
}

func (p *PinpointLogLevels) WithOverride(l slog.Level, packagePrefixes ...string) *PinpointLogLevels {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, curPrefix := range packagePrefixes {
		p.prefixes = append(p.prefixes, locationPrefixEntry{
			LocationPrefix: curPrefix,
			LogLevel:       l,
		})
	}

	p.sortPrefixes()

	return p
}

func (p *PinpointLogLevels) sortPrefixes() {
	slices.SortFunc(p.prefixes, func(a, b locationPrefixEntry) int {
		ln1 := len(a.LocationPrefix)
		ln2 := len(b.LocationPrefix)
		if ln1 != ln2 {
			return cmp.Compare(ln1, ln2)
		}
		return cmp.Compare(a.LocationPrefix, b.LocationPrefix)
	})

	p.levelCache = make(map[string]slog.Level)
}

func (p *PinpointLogLevels) WithEnvironmentOverrides() *PinpointLogLevels {
	return p.WithEnvironmentListOverrides(os.Environ())
}

func (p *PinpointLogLevels) WithEnvironmentListOverrides(env []string) *PinpointLogLevels {
	for _, e := range env {
		if !strings.HasPrefix(e, TIDBITS_ENV_PREFIX) {
			continue
		}
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 || parts[1] == "" {
			continue
		}

		logLevelName := strings.TrimPrefix(parts[0], TIDBITS_ENV_PREFIX)
		logLevelName = strings.ReplaceAll(logLevelName, "_PLUS_", "+")
		logLevelName = strings.ReplaceAll(logLevelName, "_MINUS_", "-")

		var lvl slog.Level
		err := (&lvl).UnmarshalText([]byte(logLevelName))
		if err != nil {
			panic("failed to parse the Tidbit logging override: " + parts[0] + ", err=" + err.Error())
		}

		p.WithOverride(lvl, strings.Split(parts[1], ",")...)
	}

	return p
}

func (p *PinpointLogLevels) LevelForLocation(loc string) (slog.Level, bool) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	res, ok := p.levelCache[loc]
	if ok {
		return res, true
	}

	found := false
	var lvl slog.Level
	for _, curPrefix := range p.prefixes {
		if strings.HasPrefix(loc, curPrefix.LocationPrefix) {
			lvl = curPrefix.LogLevel
			found = true
		}
	}

	return lvl, found
}

func (p *PinpointLogLevels) PrintConfig(c context.Context, l *slog.Logger) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	l.InfoContext(c, "Effective Pinpoint config", slog.Any("config", p.prefixes))
}

func (p *PinpointLogLevels) FindLevel(stackFramesToSkip int) (slog.Level, bool) {
	pc, _, _, _ := runtime.Caller(stackFramesToSkip + 1)
	forPC := runtime.FuncForPC(pc)
	funcName := forPC.Name()
	return p.LevelForLocation(funcName)
}
