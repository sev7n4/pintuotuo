# Handler Caching Optimization Guide

**Status**: Implementation Ready
**Created**: 2026-03-15
**Purpose**: Guide for integrating Redis caching into handlers

---

## 📋 Overview

This guide provides step-by-step instructions for adding caching to existing handlers to improve performance and reduce database load.

---

## 🎯 Optimization Strategy

### Priority 1: High-Traffic Reads (Quick Wins)
1. **GetProductByID** - Single product details
2. **ListProducts** - Product list with pagination
3. **SearchProducts** - Search results

### Priority 2: User Context (Medium Impact)
1. **GetCurrentUser** - Current user profile
2. **GetUserByID** - User details
3. **GetTokenBalance** - Token balance

### Priority 3: Group Data (Real-time, Lower Cache Hit)
1. **ListGroups** - Group listings
2. **GetGroupByID** - Group details
3. **GetGroupProgress** - Live group progress

---

## 🔧 Implementation Pattern

### Pattern 1: Cache-Aside Pattern (Read)

```go
// Before: Direct database read
func GetProductByID(c *gin.Context) {
    id := c.Param("id")

    // Database query
    product, err := database.GetProduct(id)
    if err != nil {
        c.JSON(errors.ErrProductNotFound.Status, errors.ErrProductNotFound)
        return
    }

    c.JSON(http.StatusOK, product)
}

// After: With caching
func GetProductByID(c *gin.Context) {
    id := c.Param("id")
    cacheKey := cache.ProductKey(id)

    // Step 1: Try cache first
    cachedProduct, cacheErr := cache.Get(c, cacheKey)
    if cacheErr == nil && cachedProduct != "" {
        var product Product
        json.Unmarshal([]byte(cachedProduct), &product)
        c.JSON(http.StatusOK, product)
        return
    }

    // Step 2: Cache miss - fetch from database
    product, err := database.GetProduct(id)
    if err != nil {
        c.JSON(errors.ErrProductNotFound.Status, errors.ErrProductNotFound)
        return
    }

    // Step 3: Store in cache for future requests
    productJSON, _ := json.Marshal(product)
    cache.Set(c, cacheKey, productJSON, cache.ProductCacheTTL)

    c.JSON(http.StatusOK, product)
}
```

**Benefits**:
- Cache hit: <5ms response time (vs 10-50ms database query)
- Cache miss: Only 1-2ms overhead
- Automatic TTL invalidation

---

## 📝 Handler-by-Handler Implementation

### 1. GetProductByID

**Current**: Direct database read
**Cache**: Product details for 1 hour
**Key**: `product:{id}`
**TTL**: 1 hour

**Changes Required**:
```go
// Add at handler start
cacheKey := cache.ProductKey(id)

// Try cache
cachedProduct, cacheErr := cache.Get(c, cacheKey)
if cacheErr == nil {
    var product Product
    json.Unmarshal([]byte(cachedProduct), &product)
    return c.JSON(http.StatusOK, product)
}

// Cache miss - continue with existing logic
// Add before response
productJSON, _ := json.Marshal(product)
cache.Set(c, cacheKey, productJSON, cache.ProductCacheTTL)
```

---

### 2. ListProducts

**Current**: Database query with pagination
**Cache**: Product list for 5 minutes
**Key**: `products:list:{status}:page:{n}:limit:{m}`
**TTL**: 5 minutes

**Changes Required**:
```go
// Build cache key
status := c.Query("status")
page := c.Query("page")
limit := c.Query("limit")
cacheKey := cache.ProductListKey(page, limit, status)

// Try cache
cachedList, cacheErr := cache.Get(c, cacheKey)
if cacheErr == nil {
    var products []Product
    json.Unmarshal([]byte(cachedList), &products)
    return c.JSON(http.StatusOK, gin.H{
        "data": products,
        "page": page,
        "limit": limit,
    })
}

// Cache miss - continue with existing logic
// Add before response
listJSON, _ := json.Marshal(products)
cache.Set(c, cacheKey, listJSON, cache.ProductListTTL)
```

---

### 3. SearchProducts

**Current**: Database full-text search
**Cache**: Search results for 10 minutes
**Key**: `products:search:{query}:page:{n}:limit:{m}`
**TTL**: 10 minutes

**Changes Required**:
```go
// Build cache key
query := c.Query("q")
page := c.Query("page")
limit := c.Query("limit")
cacheKey := cache.ProductSearchKey(query, page, limit)

// Try cache
cachedResults, cacheErr := cache.Get(c, cacheKey)
if cacheErr == nil {
    var results []Product
    json.Unmarshal([]byte(cachedResults), &results)
    return c.JSON(http.StatusOK, results)
}

// Cache miss - continue with existing logic
// Add before response
resultsJSON, _ := json.Marshal(results)
cache.Set(c, cacheKey, resultsJSON, cache.SearchResultsTTL)
```

---

### 4. GetCurrentUser

**Current**: Database query
**Cache**: User profile for 30 minutes
**Key**: `user:{id}`
**TTL**: 30 minutes

**Changes Required**:
```go
// Get user ID from context
userID := c.GetInt("user_id")
cacheKey := cache.UserKey(userID)

// Try cache
cachedUser, cacheErr := cache.Get(c, cacheKey)
if cacheErr == nil {
    var user User
    json.Unmarshal([]byte(cachedUser), &user)
    return c.JSON(http.StatusOK, user)
}

// Cache miss - continue with existing logic
// Add before response
userJSON, _ := json.Marshal(user)
cache.Set(c, cacheKey, userJSON, cache.UserCacheTTL)
```

---

### 5. GetTokenBalance

**Current**: Database query
**Cache**: Token balance for 5 minutes
**Key**: `token:balance:{user_id}`
**TTL**: 5 minutes

**Changes Required**:
```go
// Get user ID
userID := c.GetInt("user_id")
cacheKey := cache.TokenBalanceKey(userID)

// Try cache
cachedBalance, cacheErr := cache.Get(c, cacheKey)
if cacheErr == nil {
    var balance struct {
        UserID  int `json:"user_id"`
        Balance int `json:"balance"`
    }
    json.Unmarshal([]byte(cachedBalance), &balance)
    return c.JSON(http.StatusOK, balance)
}

// Cache miss - continue with existing logic
// Add before response
balanceJSON, _ := json.Marshal(balance)
cache.Set(c, cacheKey, balanceJSON, cache.TokenBalanceTTL)
```

---

## 🔄 Cache Invalidation Strategy

### Update/Delete Handlers Must Invalidate Cache

**Pattern**:
```go
func UpdateProduct(c *gin.Context) {
    // ... existing update logic ...

    // Invalidate product cache
    productID := c.Param("id")
    cache.Delete(c, cache.ProductKey(productID))

    // Invalidate product lists
    cache.InvalidatePatterns(c, "products:list:*")

    // Invalidate search results
    cache.InvalidatePatterns(c, "products:search:*")

    c.JSON(http.StatusOK, updatedProduct)
}

func DeleteProduct(c *gin.Context) {
    // ... existing delete logic ...

    // Invalidate caches
    productID := c.Param("id")
    cache.Delete(c, cache.ProductKey(productID))
    cache.InvalidatePatterns(c, "products:list:*")
    cache.InvalidatePatterns(c, "products:search:*")

    c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
```

---

## 📊 Expected Performance Impact

### Before Caching
```
GetProductByID     : 10-50ms (database query)
ListProducts       : 25-100ms (database query + pagination)
SearchProducts     : 50-200ms (full-text search)
GetTokenBalance    : 15-60ms (database query)
```

### After Caching (with 70% hit ratio)
```
GetProductByID     : 4ms average (3ms cache hit + 47ms cache miss)
ListProducts       : 15ms average (5ms cache hit + 95ms cache miss)
SearchProducts     : 35ms average (5ms cache hit + 185ms cache miss)
GetTokenBalance    : 10ms average (3ms cache hit + 57ms cache miss)
```

**Overall Improvement**: 50-60% response time reduction

---

## ✅ Implementation Checklist

- [ ] GetProductByID - Add cache-aside pattern
- [ ] ListProducts - Add cache-aside pattern
- [ ] SearchProducts - Add cache-aside pattern
- [ ] UpdateProduct - Add cache invalidation
- [ ] DeleteProduct - Add cache invalidation
- [ ] GetCurrentUser - Add cache-aside pattern
- [ ] GetUserByID - Add cache-aside pattern
- [ ] GetTokenBalance - Add cache-aside pattern
- [ ] All update handlers - Add cache invalidation
- [ ] Test with 1000+ requests - Verify cache hit ratio
- [ ] Monitor Redis memory usage - Ensure TTLs work

---

## 🐛 Debugging Cache Issues

### Check Cache Hit Rate
```go
// Add metrics
metrics.RecordCacheOperation("product", "get", duration, isHit)
```

### Clear Cache (Development Only)
```go
// Clear all products cache
cache.InvalidatePatterns(c, "product:*")
cache.InvalidatePatterns(c, "products:*")
cache.InvalidatePatterns(c, "token:balance:*")
```

### Verify Cache Content
```bash
# Connect to Redis
redis-cli

# Get cached value
GET product:123

# List all keys
KEYS product:*

# Check TTL
TTL product:123
```

---

## 📚 Related Functions

- `cache.ProductKey(id)` - Generate product cache key
- `cache.ProductListKey(page, limit, status)` - Generate product list key
- `cache.ProductSearchKey(query, page, limit)` - Generate search key
- `cache.Get(ctx, key)` - Retrieve from cache
- `cache.Set(ctx, key, value, ttl)` - Store in cache
- `cache.Delete(ctx, key)` - Remove from cache
- `cache.InvalidatePatterns(ctx, pattern)` - Pattern-based invalidation

---

**Next Step**: Implement caching in priority order, test performance, monitor metrics.
