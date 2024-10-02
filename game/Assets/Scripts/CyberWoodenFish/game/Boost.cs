using UnityEngine;

namespace CyberWoodenFish.game
{
    public class Boost: MonoBehaviour
    {
        public static void GetBoost()
        {
            PlayerPrefs.SetInt("boost", 1);
        }
    }
}