using UnityEngine;
using UnityEngine.Purchasing;
using UnityEngine.SceneManagement;
using UnityEngine.UI;

namespace CyberWoodenFish.ui
{
    public class MainUIManager : MonoBehaviour
    {
        [SerializeField] private Button button;
        [SerializeField] private IAPListener iapListener;
        // Start is called once before the first execution of Update after the MonoBehaviour is created
        void Start()
        {
            iapListener.onPurchaseComplete.AddListener(Debug.Log);
            
            button.onClick.AddListener(() =>
            {
                SceneManager.LoadScene("GameScene");
            });
        }

        // Update is called once per frame
        void Update()
        {
        
        }
    }
}
