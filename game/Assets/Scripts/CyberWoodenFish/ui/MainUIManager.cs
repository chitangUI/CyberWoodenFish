using UnityEngine;
using UnityEngine.SceneManagement;
using UnityEngine.UI;

namespace CyberWoodenFish.ui
{
    public class MainUIManager : MonoBehaviour
    {
        [SerializeField] private Button button;
        // Start is called once before the first execution of Update after the MonoBehaviour is created
        void Start()
        {
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
