func (s *BillingService) GetProviders(filter *BillingFilter) ([]string, error) {
	query := `
		SELECT DISTINCT provider 
		FROM api_usage_logs aul
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if filter != nil {
		if filter.MerchantID != nil {
			query += fmt.Sprintf(" AND aul.key_id IN (SELECT id FROM merchant_api_keys WHERE merchant_id = $%d)", argIndex)
			args = append(args, *filter.MerchantID)
			argIndex++
		}
		if filter.StartDate != nil {
			query += fmt.Sprintf(" AND aul.created_at >= $%d", argIndex)
			args = append(args, *filter.StartDate)
			argIndex++
		}
		if filter.EndDate != nil {
		 query += fmt.Sprintf(" AND aul.created_at <= $%d", argIndex)
		 args = append(args, *filter.EndDate)
            argIndex++
        }
    }

	query += " ORDER BY provider"

	rows, err := s.db.Query(query, args...)
	if err != nil {
        return nil, fmt.Errorf("failed to get providers: %w", err)
    }
    defer rows.Close()

    var providers []string
    for rows.Next() {
        var provider string
        err := rows.Scan(&provider)
        if err != nil {
            continue
        }
        providers = append(providers, provider)
    }

    return providers, nil
}

func (s *BillingService) GetModels(provider string) ([]string, error) {
    query := `
        SELECT DISTINCT model 
        FROM api_usage_logs
        WHERE 1=1
    `
    args := []interface{}{}
    argIndex := 1

    if provider != "" {
        query += fmt.Sprintf(" AND provider = $%d", argIndex)
        args = append(args, provider)
        argIndex++
    }

    query += " ORDER BY model"

    rows, err := s.db.Query(query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to get models: %w", err)
    }
    defer rows.Close()

    var models []string
    for rows.Next() {
        var model string
        err := rows.Scan(&model)
        if err != nil {
            continue
        }
        models = append(models, model)
    }

    return models, nil
}
