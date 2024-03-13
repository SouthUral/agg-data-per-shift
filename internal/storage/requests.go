package storage

const (
	getStreamOffset = `
	WITH latest_sessions AS (
		SELECT DISTINCT ON (ds.object_id)
			ds.shift_id,
			ds.object_id,
			ds.event_offset,
			ds.time_update_session
		FROM reports.drivers_sessions ds 
		ORDER BY ds.object_id, ds.time_update_session DESC
	),
	latest_shifts AS (
		SELECT DISTINCT ON (s.object_id)
			s.id,
			s.object_id,
			s.event_offset,
			s.updated_time,
			ds.time_update_session,
			ds.event_offset as session_offset
		FROM reports.shifts s
		JOIN latest_sessions ds ON s.id = ds.shift_id
		WHERE s.updated_time > (SELECT MAX(updated_time) - INTERVAL '1 hour' FROM reports.shifts)
		ORDER BY s.object_id, s.updated_time DESC
	)
	SELECT 
		coalesce(LEAST(MIN(ls.session_offset), MIN(ls.event_offset)), 0) AS min_value
	FROM latest_shifts ls
	WHERE ls.time_update_session > (SELECT MAX(time_update_session) - INTERVAL '1 hour' FROM reports.drivers_sessions)
	LIMIT 1;
	`
	// запросы в БД для сообщений от модуля aggMileageHours
	// запросы на получение последней смены по objectId
	getLastObjShift = "SELECT * FROM reports.shifts WHERE object_id = $1 ORDER BY id DESC LIMIT 1"
	// запрос на получение последней сессии по object_id
	getLastObjSession = "SELECT * FROM reports.drivers_sessions WHERE object_id = $1 ORDER BY id DESC LIMIT 1"
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
		mileage_empty, 
		mileage_gps_start, 
		mileage_gps_current, 
		mileage_gps_end, 
		mileage_gps_loaded, 
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
		mileage_empty, 
		mileage_gps_start, 
		mileage_gps_current, 
		mileage_gps_end, 
		mileage_gps_loaded, 
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
		$20
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
		mileage_empty = $14,
		mileage_gps_start = $15,
		mileage_gps_current = $16,
		mileage_gps_end = $17,
		mileage_gps_loaded = $18,
		mileage_gps_empty = $19
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
		mileage_empty = $12,
		mileage_gps_start = $13,
		mileage_gps_current = $14,
		mileage_gps_end = $15,
		mileage_gps_loaded = $16,
		mileage_gps_empty = $17
	WHERE id = $1;
	`
)
