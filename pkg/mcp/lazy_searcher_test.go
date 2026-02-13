package mcp

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/toss/apps-in-toss-ax/pkg/search"
)

// fakeSearcher는 테스트용으로 실제 HTTP 요청 없이 Searcher를 생성합니다.
func fakeSearcher() (*search.Searcher, error) {
	return search.NewTestSearcher()
}

func TestLazySearcher_SingleInit(t *testing.T) {
	var callCount atomic.Int32

	ls := newLazySearcher(func() (*search.Searcher, error) {
		callCount.Add(1)
		return fakeSearcher()
	})

	ctx := context.Background()

	s1, err := ls.get(ctx)
	if err != nil {
		t.Fatalf("First get failed: %v", err)
	}

	s2, err := ls.get(ctx)
	if err != nil {
		t.Fatalf("Second get failed: %v", err)
	}

	if s1 != s2 {
		t.Error("Expected same instance on repeated calls")
	}

	if callCount.Load() != 1 {
		t.Errorf("Expected initFn called once, got %d", callCount.Load())
	}
}

func TestLazySearcher_ConcurrentGet(t *testing.T) {
	var callCount atomic.Int32

	ls := newLazySearcher(func() (*search.Searcher, error) {
		callCount.Add(1)
		return fakeSearcher()
	})

	ctx := context.Background()
	const goroutines = 50

	var wg sync.WaitGroup
	results := make([]*search.Searcher, goroutines)
	errs := make([]error, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = ls.get(ctx)
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d failed: %v", i, err)
		}
	}

	// 모든 goroutine이 같은 인스턴스를 받아야 함
	first := results[0]
	for i, s := range results {
		if s != first {
			t.Errorf("goroutine %d got different instance", i)
		}
	}

	if callCount.Load() != 1 {
		t.Errorf("Expected initFn called once, got %d", callCount.Load())
	}
}

func TestLazySearcher_RetryOnError(t *testing.T) {
	var callCount atomic.Int32
	initErr := errors.New("init failed")

	ls := newLazySearcher(func() (*search.Searcher, error) {
		n := callCount.Add(1)
		if n <= 2 {
			return nil, initErr
		}
		return fakeSearcher()
	})

	ctx := context.Background()

	// 첫 번째, 두 번째 호출은 실패해야 함
	_, err := ls.get(ctx)
	if err == nil {
		t.Fatal("Expected error on first call")
	}

	_, err = ls.get(ctx)
	if err == nil {
		t.Fatal("Expected error on second call")
	}

	// 세 번째 호출은 성공해야 함 (재시도 가능)
	s, err := ls.get(ctx)
	if err != nil {
		t.Fatalf("Expected success on third call, got: %v", err)
	}
	if s == nil {
		t.Fatal("Expected non-nil searcher")
	}

	// 이후 호출은 캐시된 인스턴스 반환
	s2, err := ls.get(ctx)
	if err != nil {
		t.Fatalf("Expected cached success, got: %v", err)
	}
	if s != s2 {
		t.Error("Expected same cached instance")
	}

	if callCount.Load() != 3 {
		t.Errorf("Expected initFn called 3 times, got %d", callCount.Load())
	}
}

func TestLazySearcher_ConcurrentGetWithInitialFailure(t *testing.T) {
	var callCount atomic.Int32

	ls := newLazySearcher(func() (*search.Searcher, error) {
		n := callCount.Add(1)
		if n == 1 {
			return nil, errors.New("first call fails")
		}
		return fakeSearcher()
	})

	ctx := context.Background()

	// 첫 번째 호출 실패
	_, err := ls.get(ctx)
	if err == nil {
		t.Fatal("Expected error on first call")
	}

	// 동시에 여러 goroutine이 재시도
	const goroutines = 20
	var wg sync.WaitGroup
	results := make([]*search.Searcher, goroutines)
	errs := make([]error, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = ls.get(ctx)
		}(i)
	}

	wg.Wait()

	// 모든 goroutine이 성공해야 함
	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d failed: %v", i, err)
		}
	}

	// 모든 goroutine이 같은 인스턴스를 받아야 함
	first := results[0]
	for i, s := range results {
		if s != first {
			t.Errorf("goroutine %d got different instance", i)
		}
	}
}
