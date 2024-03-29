using UnityEngine;

namespace CyberWoodenFish
{
    public static class Utils
    {
        /// <param name="message">Message string to show in the toast.</param>
        public static void ShowAndroidToastMessage(string message)
        {
            var unityPlayer = new AndroidJavaClass("com.unity3d.player.UnityPlayer");
            var unityActivity = unityPlayer.GetStatic<AndroidJavaObject>("currentActivity");

            if (unityActivity == null) return;
            var toastClass = new AndroidJavaClass("android.widget.Toast");
            unityActivity.Call("runOnUiThread", new AndroidJavaRunnable(() =>
            {
                var toastObject =
                    toastClass.CallStatic<AndroidJavaObject>("makeText", unityActivity, message, 0);
                toastObject.Call("show");
            }));
        }
    }
}