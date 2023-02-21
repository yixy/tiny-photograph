
文件后缀类型：

```shell
[FileTypeExtension] jpg
```

日期解析顺序：

```shell
[DateTimeOriginal] 2015:10:03 12:16:24
[ModifyDate] 2015:10:03 12:16:24
[CreateDate] 2015:10:03 12:16:24
[FileModifyDate] 2023:02:11 23:32:16+08:00
#[SubSecDateTimeOriginal] 2014:10:03 12:16:24.00
#[SubSecModifyDate] 2015:10:03 12:16:24.00
#[SubSecCreateDate] 2015:10:03 12:16:24.00
```

## 时区参考资料

https://blog.filemeta.org/2018/11/best-practice-for-date-and-time-metadata.html

>  I have a collection of more than 100,000 family photos and short videos. The photos are all in JPEG (.jpg) format with EXIF metadata. Videos are in a mix of Audio Video Interleave (.avi), QuickTime (.mov), and MPEG-4 (.mp4) formats.

> For JPEG-EXIF images, the relevant date property is EXIF:DateTimeOriginal. According to the EXIF standard, the property is in ISO 8601 format but does not include the timezone suffix and should be rendered in local time - that is, in the time zone in which the photo was taken.

> Both MP4 and Quicktime (.mov) video files use the ISOM Format. For these files, the relevant property is creation_time which is stored internally in binary form. According the ISOM specification, creation_time and other date properties are in UTC.

> Neither of these formats include timezone information by default. EXIF defines an optional timezone property but I haven't found any file samples that include it. Many cameras don't have a timezone setting. For example, the Fuji camera I have only has a local time setting. That's no problem for JPEG files but for video files (in Quicktime .mov format) the Fuji camera fills in creation_time with the local time even though the property is supposed to be UTC. UTC is not possible because the camera doesn't have timezone information.

> My Canon camera does have a timezone setting. For photos (in JPEG format) it fills in DateTimeOriginal with the local time, as expected. For videos (in .MP4 format) it fills in creation_time in UTC. In both cases, the Canon camera includes timezone information in a proprietary Canon property as part of the Makernote.

https://photo.stackexchange.com/questions/82166/is-it-possible-to-get-the-time-a-photo-was-taken-timezone-aware

> There is the DateTimeOriginal EXIF tag, but that is in local time. And it seems there is no standard EXIF tag for the time zone. 

> Then, there is the EXIF GPS timestamp, which is thankfully in UTC. However, as established here, the GPS timestamp is for when the GPS location fix was obtained. I have an example of a photo I took after an international flight where the last GPS fix was before the flight, in another country, hours and thousands of miles away.
