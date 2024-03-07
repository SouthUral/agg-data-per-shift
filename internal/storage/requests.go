package storage

const (
	// запросы в БД для сообщений от модуля aggMileageHours
	// запросы на получение последней смены по objectId
	getLastObjShift = "SELECT * FROM reports.shifts WHERE object_id = $1 ORDER BY id LIMIT 1"
	// запрос на получение последней сессии по object_id
	getLastObjSession = "SELECT * FROM reports.drivers_sessions WHERE object_id = $1 ORDER BY id LIMIT 1"
	// запрос на добавление новой смены в БД
	addNewShift = `
	INSERT INTO reports.shifts (
		num_shift,
		object_id, 
		shift_date_start, 
		shift_date_end, 
		shift_date, 
		updated_time, 
		event_offset, 
		current_driver_id, 
		loaded, 
		eng_hours_start, 
		eng_hours_current, 
		eng_hours_end, 
		mileage_start, 
		mileage_current, 
		mileage_end, 
		mileage_loaded, 
		mileage_at_beginning_of_loading, 
		mileage_empty, 
		mileage_gps_start, 
		mileage_gps_current, 
		mileage_gps_end, 
		mileage_gps_loaded, 
		mileage_gps_at_beginning_of_loading, 
		mileage_gps_empty
	) VALUES (
		$1,
		$2,
		$3, 
		$4, 
		$5, 
		$6,
		$7, 
		$8, 
		$9, 
		$10,
		$11, 
		$12, 
		$13,
		$14,
		$15,
		$16,
		$17,
		$18, 
		$19,
		$20, 
		$21,
		$22,
		$23,
		$24
	)
	RETURNING id;`
	addNewSession = `
	INSERT INTO reports.drivers_sessions (
		shift_id,
		object_id,
		driver_id,
		event_offset,
		time_start_session,
		time_update_session,
		av_speed,
		eng_hours_start, 
		eng_hours_current, 
		eng_hours_end, 
		mileage_start, 
		mileage_current, 
		mileage_end, 
		mileage_loaded, 
		mileage_at_beginning_of_loading, 
		mileage_empty, 
		mileage_gps_start, 
		mileage_gps_current, 
		mileage_gps_end, 
		mileage_gps_loaded, 
		mileage_gps_at_beginning_of_loading, 
		mileage_gps_empty
	) VALUES (
		$1,
		$2,
		$3, 
		$4, 
		$5, 
		$6,
		$7, 
		$8, 
		$9, 
		$10,
		$11, 
		$12, 
		$13,
		$14,
		$15,
		$16,
		$17,
		$18, 
		$19,
		$20, 
		$21,
		$22
	)
	RETURNING id;`
	updateShift = `
	UPDATE reports.shifts
	SET 
		shift_date_end = $2,
		updated_time = $3,
		event_offset = $4,
		current_driver_id = $5,
		loaded = $6,
		eng_hours_start = $7,
		eng_hours_current = $8,
		eng_hours_end = $9,
		mileage_start = $10,
		mileage_current = $11,
		mileage_end = $12,
		mileage_loaded = $13,
		mileage_at_beginning_of_loading = $14,
		mileage_empty = $15,
		mileage_gps_start = $16,
		mileage_gps_current = $17,
		mileage_gps_end = $18,
		mileage_gps_loaded = $19,
		mileage_gps_at_beginning_of_loading = $20,
		mileage_gps_empty = $21
	WHERE id = $1;`
	updateDriverSession = `
	UPDATE reports.drivers_sessions
	SET 
		event_offset = $2,
		time_update_session = $3,
		av_speed = $4,
		eng_hours_start = $5,
		eng_hours_current = $6,
		eng_hours_end = $7,
		mileage_start = $8,
		mileage_current = $9,
		mileage_end = $10,
		mileage_loaded = $11,
		mileage_at_beginning_of_loading = $12,
		mileage_empty = $13,
		mileage_gps_start = $14,
		mileage_gps_current = $15,
		mileage_gps_end = $16,
		mileage_gps_loaded = $17,
		mileage_gps_at_beginning_of_loading = $18,
		mileage_gps_empty = $19
	WHERE id = $1;
	`
)
