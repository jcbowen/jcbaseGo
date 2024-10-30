package main

import (
	"github.com/jcbowen/jcbaseGo/component/attachment"
	"log"
)

func main() {
	log.SetPrefix("[test] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	strBase64 := "data:image/jpeg;base64,iVBORw0KGgoAAAANSUhEUgAAACgAAAAoCAYAAACM/rhtAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAyFpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuNi1jMTQyIDc5LjE2MDkyNCwgMjAxNy8wNy8xMy0wMTowNjozOSAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIiB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiIHhtcDpDcmVhdG9yVG9vbD0iQWRvYmUgUGhvdG9zaG9wIENDIChXaW5kb3dzKSIgeG1wTU06SW5zdGFuY2VJRD0ieG1wLmlpZDo3RTY1QkQ3M0RFREMxMUVFQjA4NUIyRTJFQzUxMzUzOCIgeG1wTU06RG9jdW1lbnRJRD0ieG1wLmRpZDo3RTY1QkQ3NERFREMxMUVFQjA4NUIyRTJFQzUxMzUzOCI+IDx4bXBNTTpEZXJpdmVkRnJvbSBzdFJlZjppbnN0YW5jZUlEPSJ4bXAuaWlkOjdFNjVCRDcxREVEQzExRUVCMDg1QjJFMkVDNTEzNTM4IiBzdFJlZjpkb2N1bWVudElEPSJ4bXAuZGlkOjdFNjVCRDcyREVEQzExRUVCMDg1QjJFMkVDNTEzNTM4Ii8+IDwvcmRmOkRlc2NyaXB0aW9uPiA8L3JkZjpSREY+IDwveDp4bXBtZXRhPiA8P3hwYWNrZXQgZW5kPSJyIj8+JaP9UgAAEBJJREFUeNqsWAl4FFW2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"

	attach := attachment.New().Upload(&attachment.Options{
		Group: "test",

		FileData: strBase64,
		FileType: "image", // 确保指定了 FileType
	})

	res := attach.SetBeforeSave(func(a *attachment.Attachment) bool {
		return true
	}).Save()
	if res.HasError() {
		log.Println("Error occurred during saving:", res.Error())
	} else {
		log.Println("FileType:", res.FileType)
		log.Println("FIleMd5:", res.FileMD5)
		log.Println("FileAttachment:", res.FileAttachment)
	}
}
