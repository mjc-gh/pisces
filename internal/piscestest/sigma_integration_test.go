package piscestest

import (
    "context"
    "path/filepath"
    "runtime"
    "testing"
    "time"

    "github.com/mjc-gh/pisces"
    "github.com/mjc-gh/pisces/engine"
)

func TestSigmaRuleMatchesTestSite(t *testing.T) {
    srv := NewTestWebServer("sigma_site")
    defer srv.Close()

    ctx, cancel := NewTestContext()
    defer cancel()

    logger := pisces.SetupLogger(true)
    _, thisFile, _, ok := runtime.Caller(0)
    if !ok {
        t.Fatal("runtime.Caller failed")
    }
    pkgDir := filepath.Dir(thisFile)
    rulesDir := filepath.Join(pkgDir, "testdata", "rules")

    if err := engine.InitSigmaEngine(rulesDir, logger); err != nil {
        t.Fatalf("InitSigmaEngine failed: %v", err)
    }

    e := engine.New(1, engine.WithLogger(pisces.Logger()))
    e.Start(ctx)

    params := map[string]any{"wait": 100}
    task := engine.NewTask(
        "analyze",
        srv.URL,
        engine.WithParams(params),
        engine.WithDeviceProperties("desktop", "large"),
        engine.WithUserAgent("desktop", "chrome"),
    )
    e.Add(task)

    go func() {
        time.Sleep(250 * time.Millisecond)
        e.Shutdown()
    }()

    ctxEval := context.Background()
    gotMatch := false

    for r := range e.Results() {
        if r.Error != nil {
            t.Fatalf("engine result error: %v", r.Error)
        }

        matches, err := engine.EvaluateSigmaResult(ctxEval, r, logger)
        if err != nil {
            t.Fatalf("EvaluateSigmaResult error: %v", err)
        }

        for _, m := range matches {
            if m.Match && m.Rule.ID == "pisces-test-title-001" {
                gotMatch = true
            }
        }
    }

    if !gotMatch {
        t.Fatalf("expected Sigma rule pisces-test-title-001 to match test page, but no match occurred")
    }
}