package core

// func ActiveDeleteExpiredKeys() {
//     for {
//         expiredCount := 0
//         sampleCountRemain := constant.ActiveExpireSampleSize

//         for key, expiredTime := range dictStore.GetExpireDictStore() {
//             sampleCountRemain--
//             if sampleCountRemain < 0 {
//                 break
//             }

//             if time.Now().UnixMilli() > int64(expiredTime) {
//                 dictStore.Del(key)
//                 expiredCount++
//             }
//         }

//         if float64(expiredCount)/float64(constant.ActiveExpireSampleSize) <= constant.ActiveExpireThreshold {
//             break
//         }
//     }
// }
