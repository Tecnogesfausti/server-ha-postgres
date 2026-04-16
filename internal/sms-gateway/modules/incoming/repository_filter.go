package incoming

import "time"

type SelectFilter struct {
	ExtID     string
	UserID    string
	DeviceID  string
	Sender    string
	Type      MessageType
	StartDate time.Time
	EndDate   time.Time
}

func (f *SelectFilter) WithExtID(extID string) *SelectFilter {
	f.ExtID = extID
	return f
}

func (f *SelectFilter) WithUserID(userID string) *SelectFilter {
	f.UserID = userID
	return f
}

func (f *SelectFilter) WithDeviceID(deviceID string) *SelectFilter {
	f.DeviceID = deviceID
	return f
}

func (f *SelectFilter) WithSender(sender string) *SelectFilter {
	f.Sender = sender
	return f
}

func (f *SelectFilter) WithType(messageType MessageType) *SelectFilter {
	f.Type = messageType
	return f
}

func (f *SelectFilter) WithDateRange(start, end time.Time) *SelectFilter {
	f.StartDate = start
	f.EndDate = end
	return f
}

type SelectOptions struct {
	Limit  int
	Offset int
}

func (o *SelectOptions) WithLimit(limit int) *SelectOptions {
	o.Limit = limit
	return o
}

func (o *SelectOptions) WithOffset(offset int) *SelectOptions {
	o.Offset = offset
	return o
}
