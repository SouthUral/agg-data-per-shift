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
)
