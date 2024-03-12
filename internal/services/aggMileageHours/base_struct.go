package aggmileagehours

type aggData struct {
	EngHoursData   *engHours    // данные по моточасам за сессию
	MileageData    *mileageData // данные по пробегу за смену
	MileageGPSData *mileageData // данные пробега по GPS за сессию
}

func (a *aggData) initAggDataFields(event *eventData) {
	a.EngHoursData = initNewEngHours(event)
	a.MileageData = initNewMileageData(event.mileage)
	a.MileageGPSData = initNewMileageData(event.gpsMileage)
}

// загрузка данных из старой структуры
func (a *aggData) initNewAggDataFields(oldEngData *engHours, oldMileageData, oldMileageGPSData *mileageData, mileage, mileageGPS int) {
	a.EngHoursData = oldEngData.createNewEngHours()
	a.MileageData = oldMileageData.createNewMileageData(mileage)
	a.MileageGPSData = oldMileageGPSData.createNewMileageData(mileageGPS)
}

// загрузка данных из интерфейсов
func (a *aggData) loadingDataFromInterface(mileageData, mileageGPSData, engData interface{}) error {
	var err error
	a.MileageData, err = initNewMileageDataLoadingDBData(mileageData)
	if err != nil {
		return err
	}
	a.MileageGPSData, err = initNewMileageDataLoadingDBData(mileageGPSData)
	if err != nil {
		return err
	}
	a.EngHoursData, err = initEngHoursLoadingDBData(engData)
	return err
}

func (a *aggData) updateDataFields(eventData *eventData, objLoaded bool) {
	a.EngHoursData.updateEngHours(eventData.engineHours)
	a.MileageData.updateMileageData(eventData.mileage, objLoaded)
	a.MileageGPSData.updateMileageData(eventData.gpsMileage, objLoaded)
}

func (a aggData) GetEngHoursData() interface{} {
	return *a.EngHoursData
}
func (a aggData) GetMileageData() interface{} {
	return *a.MileageData
}
func (a aggData) GetMileageGPSData() interface{} {
	return *a.MileageGPSData
}
